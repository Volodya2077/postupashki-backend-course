-- 29. Полное дерево сборки Guardian Greaves, Bloodthorn, Harpoon и Hurricane Pike.
-- Строка на каждый путь вниз, сам корень не выводится; lvl с 1, qty — quantity последнего
-- шага, qty_total — произведение quantity по пути, path — имена от корня через ' > '.
--
-- Колонки: root_item, lvl, component, qty, qty_total, path
-- Порядок: root_item по возрастанию, затем path по возрастанию
-- Строк:   34

WITH RECURSIVE tree AS (
    SELECT root.name AS root_item, 1 AS lvl, r.component_id, c.name AS component,
           r.quantity AS qty, r.quantity AS qty_total,
           root.name || ' > ' || c.name AS path
    FROM items root
    JOIN item_recipe r ON r.item_id = root.item_id
    JOIN items c ON c.item_id = r.component_id
    WHERE root.name IN ('Guardian Greaves', 'Bloodthorn', 'Harpoon', 'Hurricane Pike')

    UNION ALL
    SELECT tree.root_item, tree.lvl + 1, r.component_id, c.name,
           r.quantity, tree.qty_total * r.quantity,
           tree.path || ' > ' || c.name
    FROM tree
    JOIN item_recipe r ON r.item_id = tree.component_id
    JOIN items c ON c.item_id = r.component_id

)
SELECT root_item, lvl, component, qty, qty_total, path
FROM tree
ORDER BY root_item, path;
