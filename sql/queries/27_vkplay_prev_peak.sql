-- 27. Пик онлайна против пика предыдущего эфира того же стримера.
-- Только площадка 'VK Play Live'; prev_peak — пик предыдущего по started_at эфира того
-- же стримера, delta — их разность; первый эфир каждого стримера не выводится.
--
-- Колонки: streamer, followers, started_at, title, peak_viewers, prev_peak, delta
-- Порядок: followers по убыванию, внутри стримера started_at по возрастанию
-- Строк:   592

WITH s AS (
    SELECT sr.nickname AS streamer, sr.followers, st.started_at, st.title, st.peak_viewers,
           LAG(st.peak_viewers) OVER (PARTITION BY st.streamer_id ORDER BY st.started_at) AS prev_peak
    FROM platforms pl
    JOIN streamers sr ON sr.platform_id = pl.platform_id
    JOIN streams st ON st.streamer_id = sr.streamer_id

    WHERE pl.name = 'VK Play Live'
)
SELECT streamer, followers, started_at, title, peak_viewers, prev_peak,
       peak_viewers - prev_peak AS delta
FROM s
WHERE prev_peak IS NOT NULL
ORDER BY followers DESC, started_at;
