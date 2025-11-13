# go-idm
An Internet Download Manager written in Go

curl -X 'POST' \
  'http://localhost:8081/go_idm.v1.GoIDMService/CreateAccount' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "accountName": "admin",
  "password": "admin"
}'

curl -X 'POST' \
  'http://localhost:8081/go_idm.v1.GoIDMService/CreateSession' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "accountName": "admin",
  "password": "admin"
}'

eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjMxMjk2MDQsInN1YiI6MX0.KRSpzHFxcSau2XeF2WMMnzQD6-Cj03xtHJ2EEZjIUsM

curl -X 'POST' \
  'http://localhost:8081/go_idm.v1.GoIDMService/CreateDownloadTask' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjMxMjk2MDQsInN1YiI6MX0.KRSpzHFxcSau2XeF2WMMnzQD6-Cj03xtHJ2EEZjIUsM",
  "downloadType": 1,
  "url": "https://example.com/"
}'