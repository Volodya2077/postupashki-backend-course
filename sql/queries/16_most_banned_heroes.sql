-- 16. Десятка самых банимых героев.
-- bans — строки match_bans, picks — строки match_players, bans_per_pick =
-- round(bans::numeric / picks, 2); первые десять берутся после сортировки.
--
-- Колонки: hero, bans, picks, bans_per_pick
-- Порядок: bans по убыванию, при равенстве hero по возрастанию
-- Строк:   10

SELECT name AS hero,
       match_bans.bans AS bans,
       match_players.picks AS picks,
       ROUND(match_bans.bans::numeric / match_players.picks, 2) AS bans_per_pick
FROM heroes
JOIN (SELECT hero_id, COUNT(*) AS bans FROM match_bans GROUP BY hero_id) AS match_bans
  ON match_bans.hero_id = heroes.hero_id
JOIN (SELECT hero_id, COUNT(*) AS picks FROM match_players GROUP BY hero_id) AS match_players
  ON match_players.hero_id = heroes.hero_id
ORDER BY bans DESC, hero
LIMIT 10;

