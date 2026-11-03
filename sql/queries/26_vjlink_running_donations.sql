-- 26. Нарастающая сумма донатов внутри каждого стрима Вжлинка за 2021 год.
-- Стримы Вжлинка, начавшиеся в 2021 году по UTC, строка на каждый донат; running_total —
-- нарастающая сумма донатов своего стрима по sent_at_sec.
--
-- Колонки: stream_id, title, sent_at_sec, donor, amount, running_total
-- Порядок: stream_id по возрастанию, внутри стрима sent_at_sec по возрастанию
-- Строк:   192

SELECT st.stream_id,
       st.title,
       sd.sent_at_sec,
       sd.nickname AS donor,
       sd.amount,
       SUM(sd.amount) OVER (PARTITION BY sd.stream_id ORDER BY sd.sent_at_sec, sd.donation_id) AS running_total
FROM streamers sr
JOIN streams st ON st.streamer_id = sr.streamer_id
JOIN stream_donations sd ON sd.stream_id = st.stream_id

WHERE sr.nickname = 'Вжлинк' AND st.started_at >= '2021-01-01 00:00:00+00'::timestamptz AND st.started_at < '2022-01-01 00:00:00+00'::timestamptz
ORDER BY sd.stream_id, sd.sent_at_sec, sd.donation_id;
