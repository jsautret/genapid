# IFTTT setup

To process a query from [IFTTT](https://ifttt.com/), we use a **That** Webhook with the following setup.

### That
For **That** choose: **Maker Webhooks**

1. Choose *Make a web request*
2. In *URL* enter:
>http://mydomain:9110/ifttt?q={{TextField}}

Replace *mydomain.com* by your domain and *9110* by the port where you will be running genapid. Adjust the URL query parameters accordingly on your **If** and your use case.

3. Method: **POST**
4. Content Type: **application/json**
5. Body:
>{"token":"*A_SECRET_TOKEN*"}

Replace *A_SECRET_TOKEN* by a random string. It will be used as a shared secret between IFTTT and genapid, to prevent unauthorized usage of genapid.


### Predicates for IFTTT

To process the IFTTT webhook in genapid, we use the following predicates at the beginning of our pipe:

``` yaml
- name: "Incoming IFTTT request"
  pipe:
  - name: IFTTT dedicated endpoint
    match:
      string: =In.Req.URL.Path
      value: /ifttt
  - name: Body must be JSON
    body:
	  mime: application/json
      type: json
    register: body
  - name: JSON payload must contain the token set in IFTTT
    match:
      string: = R.body.payload.token
      value: A_SECRET_TOKEN
```

We can then place other predicates after that 4 ones to process the logic of our use case.
