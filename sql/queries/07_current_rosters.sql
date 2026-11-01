-- 07. Действующие составы: кто сейчас за кого играет.
-- Только незакрытые периоды (left_at IS NULL); игроки без команды не выводятся.
--
-- Колонки: nickname, country, mmr, team_name, joined_at
-- Порядок: joined_at по возрастанию, при равенстве nickname по возрастанию
-- Строк:   86
SELECT nickname, country, mmr, teams.name AS team_name, joined_at FROM team_members
JOIN players ON players.player_id = team_members.player_id
JOIN teams ON teams.team_id = team_members.team_id
WHERE left_at IS NULL
ORDER BY joined_at, nickname;
