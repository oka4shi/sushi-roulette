import json
import urllib.request

import bs4

URL = "https://www.akindo-sushiro.co.jp/menu/menu_detail/?s_id=528"

req = urllib.request.Request(URL)

with urllib.request.urlopen(req) as res:
    body = res.read()


sushiro = bs4.BeautifulSoup(body, 'html.parser')

categories = [category.get_text()
              for category in sushiro.select_one('.category-tab').children]

menu = sushiro.select_one('.swiper-wrapper').children

sushi = []
for i, element in enumerate(menu):
    sushi_in_category = []

    category_name = categories[i]

    if (not element.name) or category_name in ['期間限定', 'お持ち帰りメニュー']:
        continue

    for i, item in enumerate(element.select('.menu-item')):
        if (not item.name):
            continue

        sushi_in_category.append({
            "name": item.select_one('.menu-item__name').get_text(),
            "img_url": item.select_one('.menu-item__img')['src']
        })

    sushi.append({"category": category_name, "sushi": sushi_in_category})

with open('../json/sushiro.json', 'w') as f:
    json.dump(sushi, f, ensure_ascii=False, indent=2)
