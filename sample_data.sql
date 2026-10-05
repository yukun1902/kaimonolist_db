-- サンプルデータを入れ直す
-- 実行先データベース: kaimonolist_db
-- ※ 今入っているデータはすべて消える

-- BEGIN〜COMMIT の間は「ひとまとまり」。途中でエラーが出たら全部なかったことになる
BEGIN;

-- ① 今のデータを全部消して、自動の番号も 1 に戻す
TRUNCATE shopping_items, categories RESTART IDENTITY;

-- ② カテゴリーを先に登録する（商品が指す先を先に用意する）
INSERT INTO categories (id, name) VALUES
    (1, '食品'),
    (2, '日用品');

-- ③ 商品を登録する（category_id で categories の番号を指す）
INSERT INTO shopping_items (id, name, category_id, purchase_date, is_purchased) VALUES
    (1, '牛乳', 1, '2026-10-05', false),
    (2, '卵',   1, '2026-10-05', true),
    (3, '洗剤', 2, '2026-10-06', false);

-- ④ 番号を手で指定したので、次の自動の番号が続きから振られるように合わせる
SELECT setval('categories_id_seq', (SELECT MAX(id) FROM categories));
SELECT setval('shopping_items_id_seq', (SELECT MAX(id) FROM shopping_items));

COMMIT;

-- ⑤ 確認：2つの表をつなげて表示する
SELECT i.id, i.name AS 商品名, c.name AS カテゴリー, i.purchase_date AS 日付, i.is_purchased AS 購入済み
FROM shopping_items i
JOIN categories c ON i.category_id = c.id
ORDER BY i.id;
