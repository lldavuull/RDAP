惡意網站+駭客網站 兩個程式：
go-oidc-3/example/idtoken/go.app：惡意網站，登入openID後，將token傳到駭客網站，並向駭客網站進行查詢rdap。
rdap-master/cmd/rdap/main.go：駭客網站，收到token後向op換成userinfo，收到rdap查詢關鍵字時進行查詢，並將rdap查詢結果傳回。
