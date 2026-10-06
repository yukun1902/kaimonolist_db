-- 買い物リストアプリ：テーブル作成とサンプルデータ
-- 実行先データベース: kaimonolist_db（空のデータベースで実行する）
-- 実行方法: psql -U postgres -d kaimonolist_db -f schema.sql

-- ==================================================
-- 1. テーブル作成
-- ==================================================

-- カテゴリー一覧
-- ※ shopping_items から参照されるので、こちらを先に作る
CREATE TABLE categories (
    id   SERIAL PRIMARY KEY,           -- カテゴリーID（1, 2, 3…と自動で番号を振る）
    name VARCHAR(50) NOT NULL UNIQUE   -- カテゴリー名（空っぽ禁止・重複禁止）
);

-- 買い物リスト
CREATE TABLE shopping_items (
    id            SERIAL PRIMARY KEY,                -- 商品ID（自動で番号を振る）
    name          VARCHAR(100) NOT NULL,             -- 商品名（空っぽ禁止）
    category_id   INTEGER NOT NULL                   -- カテゴリーID（空っぽ禁止）
                  REFERENCES categories (id)         -- 外部キー：categories の id を指す
                  ON DELETE RESTRICT,                -- 商品が残っているカテゴリーは削除できない
    purchase_date DATE NOT NULL,                     -- 買う予定の日（空っぽ禁止）
    is_purchased  BOOLEAN NOT NULL DEFAULT false     -- 購入済みかどうか（最初は未購入）
);

-- ==================================================
-- 2. サンプルデータ投入
-- ==================================================

-- カテゴリーを先に登録する（商品が指す先を先に用意する）
INSERT INTO categories (name) VALUES
    ('乳製品・卵'),     -- 1
    ('日用品'),         -- 2
    ('野菜・果物'),     -- 3
    ('肉・魚'),         -- 4
    ('米・パン・麺'),   -- 5
    ('調味料'),         -- 6
    ('飲み物'),         -- 7
    ('お菓子'),         -- 8
    ('冷凍食品'),       -- 9
    ('薬・衛生用品');   -- 10

-- 商品を登録する（category_id で categories の番号を指す）
INSERT INTO shopping_items (name, category_id, purchase_date, is_purchased) VALUES
    ('牛乳', 1, '2026-10-05', false),   -- 1：乳製品・卵／未購入
    ('卵',   1, '2026-10-05', true),    -- 2：乳製品・卵／購入済み
    ('洗剤', 2, '2026-10-06', false);   -- 3：日用品／未購入

-- ==================================================
-- 3. 確認：2つの表をつなげて表示する
-- ==================================================
SELECT i.id, i.name AS 商品名, c.name AS カテゴリー, i.purchase_date AS 日付, i.is_purchased AS 購入済み
FROM shopping_items i
JOIN categories c ON i.category_id = c.id
ORDER BY i.id;
