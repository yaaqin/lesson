package curriculumsvc

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Kuota bikin room multiplayer (migrations/0014). User premium gak kepotong
// kuota; user biasa pakai users.room_quota, jatah awalnya dari
// multiplayer_config.initial_room_quota pas daftar. Tambahan jatah lewat
// permintaan ke admin (room_quota_requests).

const (
	RoomQuotaRequestPending  = "pending"
	RoomQuotaRequestApproved = "approved"
	RoomQuotaRequestRejected = "rejected"

	roomQuotaRequestMaxLen = 500
	MaxRoomQuotaGrant      = 100
	MaxRoomQuota           = 1000
)

var (
	ErrNoRoomQuota         = errors.New("kesempatan bikin room udah habis")
	ErrRoomQuotaPending    = errors.New("masih ada permintaan yang belum diproses")
	ErrInvalidQuotaRequest = errors.New("isi permintaan gak valid")
	ErrAlreadyPremium      = errors.New("user premium gak butuh tambahan kesempatan")
	ErrInvalidQuotaAmount  = errors.New("jumlah kesempatan gak valid")
	ErrQuotaRequestDecided = errors.New("permintaan udah diproses")
)

type RoomQuotaRequest struct {
	ID           string     `json:"id"`
	Message      string     `json:"message"`
	Status       string     `json:"status"`
	GrantedQuota *int       `json:"grantedQuota"`
	CreatedAt    time.Time  `json:"createdAt"`
	DecidedAt    *time.Time `json:"decidedAt"`
}

// RoomQuotaStatus: yang dilihat user di apps/web (sisa kesempatan +
// permintaan terakhirnya).
type RoomQuotaStatus struct {
	IsPremium     bool              `json:"isPremium"`
	RoomQuota     int               `json:"roomQuota"`
	LatestRequest *RoomQuotaRequest `json:"latestRequest"`
}

func (s *Service) GetRoomQuotaStatus(ctx context.Context, userID string) (*RoomQuotaStatus, error) {
	var st RoomQuotaStatus
	err := s.db.QueryRow(ctx, `SELECT is_premium, room_quota FROM users WHERE id = $1`, userID).
		Scan(&st.IsPremium, &st.RoomQuota)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	var req RoomQuotaRequest
	err = s.db.QueryRow(ctx, `
		SELECT id, message, status, granted_quota, created_at, decided_at
		FROM room_quota_requests WHERE user_id = $1
		ORDER BY created_at DESC LIMIT 1
	`, userID).Scan(&req.ID, &req.Message, &req.Status, &req.GrantedQuota, &req.CreatedAt, &req.DecidedAt)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
	case err != nil:
		return nil, err
	default:
		st.LatestRequest = &req
	}
	return &st, nil
}

// ConsumeRoomQuota: potong 1 kesempatan (atomic -- 2 request barengan gak
// bisa makai jatah yang sama). User premium gak usah manggil ini.
func (s *Service) ConsumeRoomQuota(ctx context.Context, userID string) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE users SET room_quota = room_quota - 1, updated_at = now()
		WHERE id = $1 AND room_quota > 0
	`, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNoRoomQuota
	}
	return nil
}

// RefundRoomQuota: balikin 1 kesempatan kalau room gagal kebikin setelah
// kuotanya kepotong.
func (s *Service) RefundRoomQuota(ctx context.Context, userID string) error {
	_, err := s.db.Exec(ctx, `UPDATE users SET room_quota = room_quota + 1, updated_at = now() WHERE id = $1`, userID)
	return err
}

func (s *Service) CreateRoomQuotaRequest(ctx context.Context, userID, message string) (*RoomQuotaRequest, error) {
	message = strings.TrimSpace(message)
	if message == "" || utf8.RuneCountInString(message) > roomQuotaRequestMaxLen {
		return nil, ErrInvalidQuotaRequest
	}

	var isPremium bool
	err := s.db.QueryRow(ctx, `SELECT is_premium FROM users WHERE id = $1`, userID).Scan(&isPremium)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if isPremium {
		return nil, ErrAlreadyPremium
	}

	req := RoomQuotaRequest{Message: message, Status: RoomQuotaRequestPending}
	err = s.db.QueryRow(ctx, `
		INSERT INTO room_quota_requests (user_id, message) VALUES ($1, $2)
		RETURNING id, created_at
	`, userID, message).Scan(&req.ID, &req.CreatedAt)
	if err != nil {
		// Unique index room_quota_requests_one_pending: 1 pending per user.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrRoomQuotaPending
		}
		return nil, err
	}
	return &req, nil
}

// ---------- Admin ----------

type AdminRoomQuotaRequest struct {
	RoomQuotaRequest
	User struct {
		ID          string  `json:"id"`
		Email       string  `json:"email"`
		DisplayName string  `json:"displayName"`
		Nickname    *string `json:"nickname"`
		RoomQuota   int     `json:"roomQuota"`
		IsPremium   bool    `json:"isPremium"`
	} `json:"user"`
}

type AdminRoomQuotaRequestList struct {
	Items        []AdminRoomQuotaRequest `json:"items"`
	Total        int                     `json:"total"`
	PendingCount int                     `json:"pendingCount"`
	Page         int                     `json:"page"`
	PageSize     int                     `json:"pageSize"`
}

// AdminListRoomQuotaRequests: status kosong = semua. Pending paling lama di
// atas (antrian), yang udah diproses paling baru di atas.
func (s *Service) AdminListRoomQuotaRequests(ctx context.Context, status string, page, pageSize int) (*AdminRoomQuotaRequestList, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	switch status {
	case "", RoomQuotaRequestPending, RoomQuotaRequestApproved, RoomQuotaRequestRejected:
	default:
		status = ""
	}

	res := AdminRoomQuotaRequestList{Items: []AdminRoomQuotaRequest{}, Page: page, PageSize: pageSize}
	if err := s.db.QueryRow(ctx, `
		SELECT count(*) FILTER (WHERE $1 = '' OR status = $1),
			count(*) FILTER (WHERE status = 'pending')
		FROM room_quota_requests
	`, status).Scan(&res.Total, &res.PendingCount); err != nil {
		return nil, err
	}

	rows, err := s.db.Query(ctx, `
		SELECT r.id, r.message, r.status, r.granted_quota, r.created_at, r.decided_at,
			u.id, u.email, u.display_name, u.username, u.room_quota, u.is_premium
		FROM room_quota_requests r
		JOIN users u ON u.id = r.user_id
		WHERE $1 = '' OR r.status = $1
		ORDER BY (r.status = 'pending') DESC,
			CASE WHEN r.status = 'pending' THEN r.created_at END ASC,
			COALESCE(r.decided_at, r.created_at) DESC
		LIMIT $2 OFFSET $3
	`, status, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var it AdminRoomQuotaRequest
		if err := rows.Scan(&it.ID, &it.Message, &it.Status, &it.GrantedQuota, &it.CreatedAt, &it.DecidedAt,
			&it.User.ID, &it.User.Email, &it.User.DisplayName, &it.User.Nickname, &it.User.RoomQuota, &it.User.IsPremium); err != nil {
			return nil, err
		}
		res.Items = append(res.Items, it)
	}
	return &res, rows.Err()
}

// AdminApproveRoomQuotaRequest: setujui + TAMBAH `grant` kesempatan ke kuota
// user (bukan nimpa sisa kuotanya).
func (s *Service) AdminApproveRoomQuotaRequest(ctx context.Context, requestID, adminID string, grant int) error {
	if grant < 1 || grant > MaxRoomQuotaGrant {
		return ErrInvalidQuotaAmount
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	userID, err := decideQuotaRequest(ctx, tx, requestID, adminID, RoomQuotaRequestApproved, &grant)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE users SET room_quota = LEAST(room_quota + $1, $2), updated_at = now() WHERE id = $3
	`, grant, MaxRoomQuota, userID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Service) AdminRejectRoomQuotaRequest(ctx context.Context, requestID, adminID string) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := decideQuotaRequest(ctx, tx, requestID, adminID, RoomQuotaRequestRejected, nil); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// decideQuotaRequest: cuma permintaan yang masih pending yang bisa diproses
// (2 admin klik barengan -> yang kedua dapet ErrQuotaRequestDecided).
// uuidPattern buat id dari path URL -- id ngaco jadi ErrNotFound, bukan
// error cast uuid dari Postgres.
var quotaRequestIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func decideQuotaRequest(ctx context.Context, tx pgx.Tx, requestID, adminID, status string, grant *int) (string, error) {
	if !quotaRequestIDPattern.MatchString(requestID) {
		return "", ErrNotFound
	}
	var userID string
	err := tx.QueryRow(ctx, `
		UPDATE room_quota_requests
		SET status = $1, granted_quota = $2, decided_by = $3, decided_at = now()
		WHERE id = $4 AND status = 'pending'
		RETURNING user_id
	`, status, grant, adminID, requestID).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		var exists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM room_quota_requests WHERE id = $1)`, requestID).Scan(&exists); err != nil {
			return "", err
		}
		if !exists {
			return "", ErrNotFound
		}
		return "", ErrQuotaRequestDecided
	}
	return userID, err
}

// AdminSetUserRoomQuota: set langsung sisa kesempatan bikin room 1 user.
func (s *Service) AdminSetUserRoomQuota(ctx context.Context, userID string, quota int) error {
	if quota < 0 || quota > MaxRoomQuota {
		return ErrInvalidQuotaAmount
	}
	if !quotaRequestIDPattern.MatchString(userID) {
		return ErrNotFound
	}
	tag, err := s.db.Exec(ctx, `UPDATE users SET room_quota = $1, updated_at = now() WHERE id = $2`, quota, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
