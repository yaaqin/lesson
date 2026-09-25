-- Kampus & umum sebelumnya sama-sama order_index 3 (seed-curriculum & seed-puzzle
-- ngasih index sendiri-sendiri), jadi urutannya gak pasti. Kampus lanjut dari
-- jalur SD-SMP-SMK, umum (non-formal) ditaruh paling akhir.
UPDATE tiers SET order_index = 3 WHERE code = 'kampus';
UPDATE tiers SET order_index = 4 WHERE code = 'umum';
