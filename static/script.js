const form = document.getElementById('item-form');
const dateInput = document.getElementById('date-input');
const categoryInput = document.getElementById('category-input');
const nameInput = document.getElementById('name-input');
const list = document.getElementById('list');
const emptyMessage = document.getElementById('empty-message');
const groupTemplate = document.getElementById('group-template');
const itemTemplate = document.getElementById('item-template');

// サーバーの場所
// http://localhost:8080 から開いたときはそのまま、HTMLを直接開いたときなどはサーバーのURLを付ける
const API_BASE = location.host === 'localhost:8080' ? '' : 'http://localhost:8080';

// サーバー（データベース）から読み込んだデータ
// 1件の形: { id: 1, name: "牛乳", category_id: 1, purchase_date: "2026-10-05", is_purchased: false }
let items = [];
let categories = []; // [{ id: 1, name: "乳製品・卵" }, ...]

// 日付の初期値は今日
dateInput.value = toDateString(new Date());

// 追加
form.addEventListener('submit', async (event) => {
  event.preventDefault();

  const name = nameInput.value.trim();
  const purchaseDate = dateInput.value;
  const categoryId = Number(categoryInput.value);
  if (!name || !purchaseDate || !categoryId) return;

  const item = await api('POST', '/api/items', {
    name,
    category_id: categoryId,
    purchase_date: purchaseDate,
  });
  if (!item) return;

  items.push(item);
  render();

  nameInput.value = '';
  nameInput.focus();
});

// チェックの切り替え
list.addEventListener('change', async (event) => {
  if (!event.target.classList.contains('item-checkbox')) return;

  const id = Number(event.target.closest('.item').dataset.id);
  const updated = await api('PATCH', `/api/items/${id}`, { is_purchased: event.target.checked });
  if (updated) {
    items = items.map((i) => (i.id === id ? updated : i));
  }
  // 失敗したときも描き直して、チェックを元の状態に戻す
  render();
});

// 削除
list.addEventListener('click', async (event) => {
  if (!event.target.classList.contains('delete-button')) return;

  const id = Number(event.target.closest('.item').dataset.id);
  const ok = await api('DELETE', `/api/items/${id}`);
  if (!ok) return;

  items = items.filter((i) => i.id !== id);
  render();
});

// サーバーにリクエストを送る
// 成功したら返ってきたデータ（削除のときは true）を、失敗したら null を返す
async function api(method, path, body) {
  try {
    const response = await fetch(API_BASE + path, {
      method,
      headers: body ? { 'Content-Type': 'application/json' } : {},
      body: body ? JSON.stringify(body) : undefined,
    });

    if (response.status === 204) return true;

    const data = await response.json();
    if (!response.ok) {
      alert(data.error || 'エラーが発生しました');
      return null;
    }
    return data;
  } catch (error) {
    alert('サーバーに接続できません');
    return null;
  }
}

// カテゴリーのプルダウンを作り直す（HTMLに書いてある選択肢は、先頭の「カテゴリー」以外消す）
function renderCategoryOptions() {
  categoryInput.querySelectorAll('option[value]:not([value=""])').forEach((option) => option.remove());
  for (const category of categories) {
    const option = document.createElement('option');
    option.value = category.id;
    option.textContent = category.name;
    categoryInput.appendChild(option);
  }
}

// items の内容を画面に描き直す
function render() {
  list.innerHTML = '';
  emptyMessage.hidden = items.length > 0;

  // 日付ごとにまとめる（日付の古い順）
  const dates = [...new Set(items.map((i) => i.purchase_date))].sort();

  for (const date of dates) {
    const group = groupTemplate.content.cloneNode(true);
    group.querySelector('.date-heading').textContent = formatDate(date);
    const ul = group.querySelector('.item-list');

    for (const item of items.filter((i) => i.purchase_date === date)) {
      const li = itemTemplate.content.cloneNode(true);
      const row = li.querySelector('.item');
      row.dataset.id = item.id;
      row.classList.toggle('checked', item.is_purchased);
      li.querySelector('.item-checkbox').checked = item.is_purchased;
      li.querySelector('.item-name').textContent = item.name;
      li.querySelector('.item-category').textContent = categoryName(item.category_id);
      ul.appendChild(li);
    }

    list.appendChild(group);
  }
}

// カテゴリーID → カテゴリー名
function categoryName(categoryId) {
  const category = categories.find((c) => c.id === categoryId);
  return category ? category.name : '';
}

// Date → "2026-10-04"
function toDateString(date) {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, '0');
  const d = String(date.getDate()).padStart(2, '0');
  return `${y}-${m}-${d}`;
}

// "2026-10-04" → "10月4日（日）"
function formatDate(dateString) {
  const [y, m, d] = dateString.split('-').map(Number);
  const week = ['日', '月', '火', '水', '木', '金', '土'][new Date(y, m - 1, d).getDay()];
  return `${m}月${d}日（${week}）`;
}

// 最初にカテゴリーと買い物リストをサーバーから読み込む
async function init() {
  const loaded = await api('GET', '/api/categories');
  if (loaded) {
    categories = loaded;
    renderCategoryOptions();
  } else {
    // サーバーにつながらないときは、HTMLに書いてある選択肢をそのまま使う
    categories = [...categoryInput.querySelectorAll('option:not([value=""])')].map((option) => ({
      id: Number(option.value),
      name: option.textContent,
    }));
    render();
    return; // 一覧も読み込めないので、ここで終わる（エラー表示を2回出さないため）
  }
  items = (await api('GET', '/api/items')) || [];
  render();
}

init();
