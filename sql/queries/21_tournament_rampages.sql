-- 21. Турнирные выступления с KDA выше шестикратного среднего.
-- KDA = (kills + assists)::numeric / GREATEST(deaths, 1); порог — строго больше шести
-- средних KDA по всей match_players; только матчи турниров; kda округлить до 2 знаков.
--
-- Колонки: match_id, tournament, player, hero, kda
-- Порядок: kda по убыванию, при равенстве match_id, затем slot
-- Строк:   60

SELECT mp.match_id,
       t.name AS tournament,
       p.nickname AS player,
       h.name AS hero,
       ROUND((mp.kills + mp.assists)::numeric / GREATEST(mp.deaths, 1), 2) AS kda
FROM match_players mp
JOIN matches m ON m.match_id = mp.match_id
JOIN tournaments t ON t.tournament_id = m.tournament_id
JOIN players p ON p.player_id = mp.player_id
JOIN heroes h ON h.hero_id = mp.hero_id
WHERE (mp.kills + mp.assists)::numeric / GREATEST(mp.deaths, 1) > 6 * (SELECT AVG((kills + assists)::numeric / GREATEST(deaths, 1)) FROM match_players)
ORDER BY kda DESC, mp.match_id, mp.slot;
