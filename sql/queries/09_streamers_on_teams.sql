-- 09. Кто из стримеров за какую команду выступает сейчас.
-- Только действующие составы; стримеры без player_id не выводятся.
--
-- Колонки: streamer, platform, player, team, tag, followers
-- Порядок: followers по убыванию, при равенстве streamer по возрастанию
-- Строк:   16

SELECT streamers.nickname AS streamer, platforms.name AS platform, players.nickname AS player, teams.name AS team, teams.tag, followers FROM streamers 
JOIN players ON players.player_id = streamers.player_id  
LEFT JOIN team_members ON team_members.player_id = players.player_id AND team_members.left_at IS NULL
LEFT JOIN teams ON teams.team_id = team_members.team_id
LEFT JOIN platforms ON platforms.platform_id = streamers.platform_id
ORDER BY followers DESC, streamer
LIMIT 16;


