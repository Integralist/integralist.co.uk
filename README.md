# integralist.co.uk

Website built using [Go](https://golang.org/), [Hugo](https://gohugo.io/) and deployed via [Netlify](https://www.netlify.com/).

## New Post

- `hugo new ./posts/`

- `vim **<Tab>`(using [fzf](https://github.com/junegunn/fzf) for file searching)

- `hugo server` (for previewing)

## DNS

* Hostname: www
* Type: CNAME
* Value: dreamy-wing-b0b998.netlify.com.

---

* Hostname: @
* Type: A
* Value: 104.198.14.52

---

* Hostname: www
* Type: A
* Value: 104.198.14.52

---

```
$ dig www.integralist.co.uk

;; ANSWER SECTION:
www.integralist.co.uk.  14399   IN      CNAME   dreamy-wing-b0b998.netlify.com.
dreamy-wing-b0b998.netlify.com. 19 IN   A       54.229.14.125

$ dig integralist.co.uk

;; ANSWER SECTION:
integralist.co.uk.      14224   IN      A       104.198.14.52

$ dig A integralist.co.uk @ns.123-reg.co.uk. +short
104.198.14.52

$ dig A integralist.co.uk @8.8.8.8 +short
104.198.14.52
```
