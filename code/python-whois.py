#需先安裝 pip install python-whois
import whois
import pandas as pd
from datetime import datetime
#requests用於發送 HTTP 請求。它的設計目標是使 HTTP 請求變得簡單易用。
import requests
#pprint 是 Python 的一個內建模組，全名為 "pretty printer"，
#用於提供更加美觀的資料輸出格式。它可以將 Python 的資料結構如列表、
#字典等以更加整齊、格式化的方式輸出，特別適合用於輸出複雜的資料結構。
from pprint import pprint

# print('網域回傳：',requests.get('https://pythonviz.com').status_code,'\n\n')
# print('ICANN 註冊資料：')
# pprint(whois.whois('pythonviz.com'),width=1)

#輸入網域
domains = [
    'pythonviz.com',
    'porkbun.com',
    'cbna.com',
    'cbnb.com',
    'cbnc.com',
    'cbnd.com',
    'cbne.com',
    'cbnf.com',
    'cbng.com',
    'cbnh.com',
    'cbni.com',
    'cbnj.com',
    'cbnk.com'
    ]

#顯示網域狀態
def domainStatus(domain):
    code = requests.get('https://' + domain).status_code
    if code == 200:
        state = '🟢'
    else:
        state = '🔴'
    return state + ' ' + str(code)

def getWHOISData(domains:list):
    #接收網域查詢的結果
    for d in domains:
        print(whois.whois(d))
        # if whois.whois(d).status == 'ok':
        #     print(d,': ok')
        # else:
        #     print(d,': error')
    #whois_response = [whois.whois(d) for d in domains]
    #將結果轉換為 pandas.DataFrame
    # 表格 = pd.DataFrame.from_records(whois_response)
    # print(表格)
    # 表格['live'] = datetime.now() - 表格['creation_date']
    # 表格['remaining'] = 表格['expiration_date'] - datetime.now()
    # for x in ['live','remaining']:
    #     表格[x] = 表格[x].astype('timedelta64[D]').astype(int)
    # 表格['expiration_date'] = 表格['expiration_date'].dt.strftime('%Y-%m-%d')
    # 表格 = 表格[['domain_name','registrar','live','expiration_date','remaining']]
    # 表格.columns = ['Domain','Registrar','Live Since','Expiration','Remaining']
    # 表格['Status'] = 表格.apply(lambda x: domainStatus(x['Domain']), axis=1)
    # 表格.sort_values(by=['Remaining'],inplace=True)
    # return 表格

pprint(getWHOISData(domains))