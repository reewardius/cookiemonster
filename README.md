# :cookie: CookieMonster
CookieMonster is a command-line tool and API for decoding and modifying vulnerable session cookies from several different frameworks. It is designed to run in automation pipelines which must be able to efficiently process a large amount of these cookies to quickly discover vulnerabilities. Additionally, CookieMonster is extensible and can easily support new cookie formats.

It's worth emphasizing that CookieMonster finds vulnerabilities in users of frameworks, usually not in the frameworks themselves. These users can resolve vulnerabilities found via CookieMonster by configuring the framework to use a strong secret key.

## Features
* Decodes and unsigns session cookies from Laravel, Django, Flask, Rack, and Express, and also handles raw JWTs.
* Rapidly evaluates cookies; ignores invalid and unsupported cookies, and quickly tests those that it can.
* Takes full advantage of Go's fast, native implementations for hash functions.
* Intelligently decodes URL-encoded and Base64-encoded cookies (i.e. the Base64 of a JWT) when the initial decoding fails.
* Supports many algorithms for HMAC-based decoders, even if the framework typically only uses one.
* Flexible base64-encoded wordlist format allows any sequence of bytes key to be added as an entry; ships with a reasonable default list.

| Framework               | Supported | Notes                                                    |
|-------------------------|-----------|----------------------------------------------------------|
| JSON Web Tokens         | ✅         | HS256, HS384, HS512                                     |
| Django                  | ✅         | Common algorithms                                       |
| Flask                   | ✅         | Common algorithms                                       |
| Rack                    | ✅         | Common algorithms                                       |
| Express (cookie-signer) | ✅         | Common algorithms                                       |
| Laravel                 | ✅         | AES-CBC-128/256 (GCM not yet supported)                 |
| itsdangerous            | ✅         | URLSafeSerializer/URLSafeTimedSerializer (default salt) |
| Others                  | ❌         | Not yet!                                                |

## Getting Started
To install CookieMonster, install Go and then install the CLI:

```bash
go install github.com/iangcarroll/cookiemonster/cmd/cookiemonster@latest
```

CookieMonster only needs two essentials: a cookie to try and unsign, and a wordlist to use. If you don't have a wordlist, CookieMonster ships with a default wordlist from the [Flask-Unsign](https://github.com/Paradoxis/Flask-Unsign) project. CookieMonster wordlists are a bit different; each line must be encoded with base64. This is because Python projects are especially liberal with inserting garbage bytes into these keys, and we need to be able to properly handle them.

An example of using the CLI with a static cookie, or with a URL:

```bash
% ./cookiemonster -cookie "gAJ9cQFYCgAAAHRlc3Rjb29raWVxAlgGAAAAd29ya2VkcQNzLg:1mgnkC:z5yDxzI06qYVAU3bkLaWYpADT4I"
🍪 CookieMonster 1.3.0
ℹ️ CookieMonster loaded the default wordlist; it has 38919 entries.
✅ Success! I discovered the key for this cookie with the django decoder; it is "changeme".

% ./cookiemonster -url "https://httpbingo.org/cookies/set?abc=gAJ9cQFYCgAAAHRlc3Rjb29raWVxAlgGAAAAd29ya2VkcQNzLg:1mgnkC:z5yDxzI06qYVAU3bkLaWYpADT4I"
🍪 CookieMonster 1.3.0
⚠️  I got a non-200 status code from this URL; it was 302.
ℹ️ CookieMonster loaded the default wordlist; it has 38919 entries.
✅ Success! I discovered the key for this cookie with the django decoder; it is "changeme".
```

```bash
echo "http://example.com" | nuclei -t cookie-extractor.yaml  | cut -d "=" -f 2 | cut -d ";" -f 1 > cookies && for cookie in $(cat cookies); do ./cookiemonster -cookie $cookie; done
```
![image](https://github.com/reewardius/cookiemonster/assets/68978608/b4b66e0c-c622-462e-962e-0a8f680c6fe9)
```bash
while read target; do
  echo "[*] Checking: $target"
  findings=$(echo "$target" | nuclei -t cookie-extractor.yaml)
  if [ -n "$findings" ]; then
    echo "$findings" | cut -d "=" -f 2 | cut -d ";" -f 1 | while read cookie; do
      echo "[+] Target: $target | Cookie: $cookie"
      ./cookiemonster -cookie "$cookie"
    done
  fi
done < targets.txt
```

## Express support
CookieMonster is capable of supporting cookies signed with `cookie-session`, which is common with Express. However, it does several strange things that require care in order to use this tool. A common response from a `cookie-session` application looks like this:

```http
set-cookie: session=eyJhbmltYWxzIjoibGlvbiJ9
set-cookie: session.sig=Vf2INocdJIqKWVfYGhXwPhQZNFI
```

In order to pass this into CookieMonster, you must include both the cookie name and the signature cookie. In this example, you would call CookieMonster like this: `cookiemonster -cookie session=eyJhbmltYWxzIjoibGlvbiJ9^Vf2INocdJIqKWVfYGhXwPhQZNFI` (note the delimiting `^` and the prefixed cookie name). The API accepts this same format in `monster.NewCookie`.

## Resigning support
CookieMonster has limited support for resigning a cookie once it has been unsigned, with the `-resign` flag. This involves modifying the body of the cookie to match your input, and then re-computing the signature with the key we discovered. Currently, you can do this for Django-decoded cookies; ensure you pass the original cookie to `-cookie`, and pass `-resign` an unencoded string of text you'd like to be inside the cookie. CookieMonster will correctly encode your input and then resign the cookie.

## API usage
CookieMonster exposes `pkg/monster`, which allows other applications to easily take advantage of it. This is much more performant than booting the CLI if you are testing many cookies. An example usage of it is below.

```go
import (
    "github.com/iangcarroll/cookiemonster/pkg/monster"
    "errors"
)

var (
	//go:embed wordlists/my-wordlist.txt
	monsterWordlist string

	wl = monster.NewWordlist()
)

func init() {
	if err := wl.LoadFromString(monsterWordlist); err != nil {
        panic(err)
    }
}

func MonsterRun(cookie string) (success bool, err error) {
	c := monster.NewCookie(cookie)

	if !c.Decode() {
		return false, errors.New("could not decode")
	}

	if _, success := c.Unsign(wl, 100); !success {
		return false, errors.New("could not unsign")
	}

	return true, nil
}
```


## Credits
CookieMonster is built with inspiration from several sources, and ships with the excellent Flask-Unsign wordlists.

* https://github.com/Paradoxis/Flask-Unsign
* https://github.com/nicksanders/rust-django-signing
