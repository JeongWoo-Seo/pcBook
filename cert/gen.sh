rm *.pem
openssl req -x509 -newkey rsa:4096 -days 365 -nodes -keyout ca-key.pem -out ca-cert.pem -subj "/C = KR/ST = Soeul/L = GwangJin/O = Internet Widgits Pty Ltd/CN = sjw/emailAddress = sjwoo6500@gmail.com"

echo "CA's self-signed certificate"
openssl x509 -in ca-cert.pem -noout -text

openssl req -newkey rsa:4096 -nodes -keyout server-key.pem -out server-req.pem -subj "/C = KR/ST = Seoul/L = GwangJin/O = PCBook Server/CN = localhost/emailAddress = server@pcbook.com"

openssl x509 -req -in server-req.pem -days 60 -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out server-cert.pem -extfile server-ext.cnf

openssl x509 -in server-cert.pem -noout -text

openssl req -newkey rsa:4096 -nodes -keyout client-key.pem -out client-req.pem -subj "/C = KR/ST = Seoul/L = GwangJin/O = PC client/CN = tester/emailAddress = client@pcbook.com"

openssl x509 -req -in client-req.pem -days 60 -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out client-cert.pem -extfile client-ext.cnf

openssl x509 -in client-cert.pem -noout -text