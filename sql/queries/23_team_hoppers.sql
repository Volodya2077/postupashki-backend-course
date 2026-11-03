-- 23. Игроки, поигравшие больше чем за одну команду.
-- Считаются различные team_id по всем периодам, строго больше одного; current_team —
-- название команды с пустым left_at, current_team и streamer могут быть NULL.
--
-- Колонки: player, country, teams_count, current_team, streamer
-- Порядок: teams_count по убыванию, при равенстве player по возрастанию
-- Строк:   25

SELECT p.nickname AS player,
       p.country,
       h.teams_count,
       ct.name AS current_team,
       s.nickname AS streamer
FROM (
    SELECT player_id, COUNT(DISTINCT team_id) AS teams_count
    FROM team_members
    GROUP BY player_id
    HAVING COUNT(DISTINCT team_id) > 1
) h
JOIN players p ON p.player_id = h.player_id
LEFT JOIN team_members cur ON cur.player_id = h.player_id AND cur.left_at IS NULL
LEFT JOIN teams ct ON ct.team_id = cur.team_id
LEFT JOIN streamers s ON s.player_id = h.player_id
ORDER BY h.teams_count DESC, p.nickname;
