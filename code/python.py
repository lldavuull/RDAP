#time 是 Python 的一個內建模組，提供了各種與時間相關的功能。
#例如，你可以使用 time 模組來獲取當前時間，或者讓程式暫停一段時間。
import time
#requests用於發送 HTTP 請求。它的設計目標是使 HTTP 請求變得簡單易用。
import requests
#BeautifulSoup 用於解析 HTML 和 XML 文件，並進行網頁爬蟲。
#它提供了一種簡單的方式來導航、搜索和修改解析樹。
from bs4 import BeautifulSoup 

table = ['a','b','c','d','e','f','g','h','i','j','k','l','m',
     'n','o','p','q','r','s','t','u','v','w','x','y','z']
url = 'https://www.twnic.tw/whois_n.cgi?query='
name = 'cbn'
k=0
for s in table:
    print("第",k,"個域名:",table[k])
    #發送 GET 請求
    #第一個為：https://www.twnic.tw/whois_n.cgi?query=cbna.com
    request = requests.get(url+name+s+".com") 
    # 輸出 HTTP 狀態碼
    print(request.status_code)
    #解析 HTML，"html.parser" 是 Python 的內建 HTML 解析器，用於解析 HTML 文檔。
    soup = BeautifulSoup(request.text,"html.parser") 
    #選擇pre標籤，<pre> 是 HTML 中的一種標籤，用於顯示預先格式化(preformatted text)的文本。
    #預先格式化：原始文本已經排版好了，例如換行，而想要保留原始排版則會用<pre>。
    #與之相對的是不保留原始文本排版的 <p> 標籤，用於顯示段落文本。
    #在 <pre> 標籤中的文本通常會保留空格和換行符，並使用等寬字體來顯示。
    #如果沒有使用 <pre> 標籤，則會使用一般字體來顯示文本，空格和換行符會被忽略。
    #<pre> 屬於「容器元素」，需要有「起始標籤」以及「結束標籤」，也因為是容器元素，裡頭可以放入其他的子元素。
    #<pre> 的顯示類型為「block 塊級元素」，預設會強制換行。
    pre = soup.select("pre")
    print(pre)
    #無匹配域名時的回應
    no_match = 'No match for domain'
    #取得pre標籤的長度
    length=len(pre)
    if length > 0:
        #取得pre標籤的第一個標籤的文字
        if pre[0].text[1:20]==no_match:
            print(name+s+".com")
    else:
        time.sleep(60)
    k=k+1