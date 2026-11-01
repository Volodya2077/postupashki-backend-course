-- 28. Три самых дорогих предмета в каждой категории лавки.
-- Категория: 'recipe' при is_recipe, иначе 'assembled' при строках в item_recipe.item_id,
-- иначе 'basic'; place 1-3 по cost убыв., при равной цене по name; purchases — строк match_items, иначе 0.
--
-- Колонки: category, place, name, cost, purchases
-- Порядок: category по возрастанию, затем place по возрастанию
-- Строк:   9

WITH classified AS (
    SELECT i.name, i.cost,
           CASE 
            WHEN i.is_recipe THEN 'recipe'
            WHEN i.item_id IN (SELECT item_id FROM item_recipe) THEN 'assembled'
            ELSE 'basic'
           END AS category,
           (SELECT COUNT(*) FROM match_items mi WHERE mi.item_id = i.item_id) AS purchases
    FROM items i
),
ranked AS (
    SELECT category, name, cost, purchases,
           ROW_NUMBER() OVER (PARTITION BY category ORDER BY cost DESC, name) AS place
    FROM classified
)
SELECT category, place, name, cost, purchases
FROM ranked
WHERE place <= 3
ORDER BY category, place;
