-- 08. Сколько в команде игроков и сколько среди них стримеров.
-- Считаются только незакрытые периоды; команда без состава остаётся
-- в ответе с нулями, стримеры без player_id не учитываются.
--
-- Колонки: name, tag, players_now, streamers_now
-- Порядок: players_now по убыванию, при равенстве name по возрастанию
-- Строк:   24

SELECT teams.name, teams.tag, COUNT(team_members.player_id) players_now, COUNT(streamers.streamer_id)  streamers_now FROM teams
LEFT JOIN team_members ON team_members.team_id = teams.team_id AND team_members.left_at IS NULL
LEFT JOIN streamers ON streamers.player_id = team_members.player_id
GROUP BY teams.name, teams.tag
ORDER BY players_now DESC, name;

