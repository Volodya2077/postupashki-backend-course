-- 30. Стоимость сборки из базовых составляющих против собственной цены предмета.
-- Берутся предметы со строками в item_recipe, сборка раскрывается до базовых (тех, у кого
-- строк в item_recipe нет, свитки тоже), количества по пути перемножаются; diff = parts_cost - cost.
--
-- Колонки: name, cost, parts_cost, base_parts, diff
-- Порядок: diff по убыванию, при равенстве name по возрастанию
-- Строк:   77

WITH RECURSIVE tree AS (
    SELECT item_id AS root_id, component_id, quantity AS qty_total
    FROM item_recipe
    UNION ALL
    SELECT tree.root_id, r.component_id, tree.qty_total * r.quantity
    FROM tree
    JOIN item_recipe r ON r.item_id = tree.component_id
)
SELECT root.name,
       root.cost,
       SUM(leaf.cost * tree.qty_total) AS parts_cost,
       SUM(tree.qty_total) AS base_parts,
       SUM(leaf.cost * tree.qty_total) - root.cost AS diff
FROM tree
JOIN items root ON root.item_id = tree.root_id
JOIN items leaf ON leaf.item_id = tree.component_id
WHERE tree.component_id NOT IN (SELECT item_id FROM item_recipe)
GROUP BY root.name, root.cost
ORDER BY diff DESC, root.name;
