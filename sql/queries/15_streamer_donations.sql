-- 15. Сколько донатов собрал каждый стример.
-- В ответе все стримеры: без донатов donations = 0 и total_amount = 0.00;
-- team — команда с left_at IS NULL, иначе пусто.
--
-- Колонки: streamer, platform, team, donations, total_amount
-- Порядок: total_amount по убыванию, при равенстве streamer по возрастанию
-- Строк:   40

SELECT s.nickname AS streamer, platforms.name AS platform, teams.name AS team,  COUNT(stream_donations.donation_id) AS donations, COALESCE(SUM(stream_donations.amount), 0) AS total_amount FROM streamers s
JOIN platforms ON platforms.platform_id = s.platform_id
LEFT JOIN streams ON streams.streamer_id = s.streamer_id
LEFT JOIN stream_donations ON stream_donations.stream_id = streams.stream_id
LEFT JOIN team_members ON team_members.player_id = s.player_id AND team_members.left_at IS NULL
LEFT JOIN teams ON teams.team_id = team_members.team_id
GROUP BY s.nickname, platforms.name, teams.name
ORDER BY total_amount DESC, streamer;




