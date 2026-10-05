-- 買い物リスト用のテーブルを作成する
-- 実行先データベース: kaimonolist_db

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

-- 最初から使うカテゴリーを登録しておく
INSERT INTO categories (name) VALUES
    ('食品'),
    ('日用品');
