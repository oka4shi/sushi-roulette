import re
import json
import urllib.request

import bs4

URL = "https://www.hama-sushi.co.jp/menu/"

req = urllib.request.Request(URL)

with urllib.request.urlopen(req) as res:
    body = res.read()


hama = bs4.BeautifulSoup(body, 'html.parser')

hama.body
sections = hama.select('section.men-section > div')

sushi = []

for section in sections:
    category_name = section.select_one('h2 > picture > img')['alt']
    if category_name not in ['はまっこセット', '地域限定', 'お持ち帰り']:
        items = section.select('div.men-products-item')

        sushi_in_category = []
        for item in items:
            text1 = item.select_one('div.men-products-item__text')
            if text1:
                name = ""

                text1 = ''.join([text for text in text1.get_text(
                    ',').split(',') if '円' not in text])

                text2 = item.select_one('div.men-products-item__small')
                if text2:
                    text2 = ''.join([text for text in text2.get_text(
                        ',').split(',') if re.match(r'（.*）', text)])
                    name = f"{text1}{text2}"
                else:
                    name = text1

                img_url = item.select_one(
                    'div.men-products-item__thumb > img')['data-src']
                sushi_in_category.append({
                    "name": name,
                    "img_url": img_url,
                })

        sushi.append({"category": category_name, "sushi": sushi_in_category})

with open('../json/hama-sushi.json', 'w') as f:
    json.dump(sushi, f, ensure_ascii=False, indent=2)
