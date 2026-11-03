-- 10. Все роли героев, которые умеют быть керри.
-- Только герои с ролью 'carry'; у каждого выводятся все его роли, по строке на роль.
--
-- Колонки: name, primary_attr, attack_type, role
-- Порядок: name по возрастанию, внутри героя role по возрастанию
-- Строк:   47
SELECT heroes.name, heroes.primary_attr, heroes.attack_type, hero_roles.role FROM heroes
JOIN hero_roles ON hero_roles.hero_id = heroes.hero_id 
WHERE heroes.hero_id IN (SELECT hero_id FROM hero_roles WHERE role = 'carry')
ORDER BY name, role;