INSERT INTO rooms (id, name, description, capacity) VALUES 
('11111111-1111-1111-1111-111111111111', 'Room A', 'With marker board', 5),
('22222222-2222-2222-2222-222222222222', 'Hall', 'For large meetings', 20),
('33333333-3333-3333-3333-333333333333', 'Calling Room', 'For 1-on-1 calls', 1)
ON CONFLICT DO NOTHING;

INSERT INTO schedules (room_id, days_of_week, start_time, end_time) VALUES 
('11111111-1111-1111-1111-111111111111', '{1,2,3,4,5}', '10:00:00', '19:00:00'),
('22222222-2222-2222-2222-222222222222', '{1,2,3,4,5,6,7}', '09:00:00', '21:00:00'),
('33333333-3333-3333-3333-333333333333', '{1,2,3,4,5}', '08:00:00', '20:00:00')
ON CONFLICT DO NOTHING;