import json
import urllib.request

import bs4

URL = "https://www.uobei.info/menu/"

req = urllib.request.Request(URL)

with urllib.request.urlopen(req) as res:
    body = res.read()


uobei = bs4.BeautifulSoup(body, 'html.parser')

menu = uobei.select_one('#normal_menu').children

sushi = []
category_name = ""
for element in menu:
    if element.name == "h2":
        category_name = tuple(element.strings)[0]
        continue

    if not element.name == "ul":
        continue

    if category_name not in ['フェア商品', 'お持ち帰り']:
        sushi_in_category = []
        for item in element:
            if not item.name:
                continue

            sushi_in_category.append({
                "name": item.select_one('p.name').get_text(),
                "img_url": item.select_one('a.luminous')['href']
            })

        sushi.append({"category": category_name, "sushi": sushi_in_category})

with open('../json/uobei.json', 'w') as f:
    json.dump(sushi, f, ensure_ascii=False, indent=2)
