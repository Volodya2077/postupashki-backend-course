-- 24. Турниры, чей призовой фонд больше суммы призовых всех предыдущих.
-- Предыдущие — со строго меньшим started_at, сравнение строгое; у первого турнира
-- prev_total и prev_count нули, не NULL; matches_played — число матчей турнира.
--
-- Колонки: tournament, started_at, prize_pool, prev_total, prev_count, matches_played
-- Порядок: started_at по возрастанию
-- Строк:   2

WITH match_counts AS (
    SELECT tournament_id, COUNT(*) AS cnt
    FROM matches
    GROUP BY tournament_id
),
stats AS (
    SELECT t.name AS tournament,
           t.started_at,
           t.prize_pool,
           COALESCE(SUM(t.prize_pool) OVER w, 0) AS prev_total,
           COUNT(*) OVER w AS prev_count,
           COALESCE(mc.cnt, 0) AS matches_played
    FROM tournaments t
    LEFT JOIN match_counts mc ON mc.tournament_id = t.tournament_id
    WINDOW w AS (ORDER BY t.started_at GROUPS BETWEEN UNBOUNDED PRECEDING AND 1 PRECEDING)
)
SELECT * FROM stats
WHERE prize_pool > prev_total
ORDER BY started_at;
