-- 28. Три самых дорогих предмета в каждой категории лавки.
-- Категория: 'recipe' при is_recipe, иначе 'assembled' при строках в item_recipe.item_id,
-- иначе 'basic'; place 1-3 по cost убыв., при равной цене по name; purchases — строк match_items, иначе 0.
--
-- Колонки: category, place, name, cost, purchases
-- Порядок: category по возрастанию, затем place по возрастанию
-- Строк:   9

WITH classified AS (
    SELECT i.item_id, i.name, i.cost,
           CASE
            WHEN i.is_recipe THEN 'recipe'
            WHEN i.item_id IN (SELECT item_id FROM item_recipe) THEN 'assembled'
            ELSE 'basic'
           END AS category
    FROM items i
),
ranked AS (
    SELECT item_id, category, name, cost,
           ROW_NUMBER() OVER (PARTITION BY category ORDER BY cost DESC, name) AS place
    FROM classified
),
purchase_count AS (
    SELECT item_id, COUNT(*) AS cnt
    FROM match_items
    GROUP BY item_id
)
SELECT r.category, r.place, r.name, r.cost, COALESCE(pc.cnt, 0) AS purchases
FROM ranked r
LEFT JOIN purchase_count pc ON pc.item_id = r.item_id
WHERE r.place <= 3
ORDER BY r.category, r.place;

