matrix-informant
================

A quick and dirty webhook service for the Matrix federated message store.

# Dependencies

- go (1.20+)
- libolm

# Usage

```bash
make
# edit ./conf/informant.sample.yaml and save as ./conf/informant.yaml
./matrix-informant
```

## Clients

### cURL

```bash
export IMG=$(base64 --wrap 0 /path/to/image.png)
curl http://localhost:9999 -X POST -H 'Content-Type: application/json' \
    -d "{\"channel\":\"\!roomid:home.server\", \"contentbytes\":\"$IMG\", \"contenttype\": \"image/png\", \"message\":\"This is a test.\", \"psk\": \"YourInformantPSKhere\"}"
```

### Python

```python
import base64
import requests

informant = 'http://localhost:9999'
channel = '!roomid:home.server'
psk = 'YourInformantPSKhere'

def inform(data, kind, message):
    requests.post(informant, json={
        'channel': channel,
        'contenttype': kind,
        'contentbytes': base64.b64encode(data).decode('utf8'),
        'message': message,
        'psk': psk
    })

with open('/path/to/image.png', 'rb') as f:
    inform(f.read(), 'image/png', 'This is a test.')
```
