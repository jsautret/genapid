# Github Webhook

This example shows how to receive [events from
Github](https://docs.github.com/en/developers/webhooks-and-events/webhooks)
and send notifications using
[Pushbullet](https://docs.pushbullet.com/) or a Google Home. You can also keep a local mirror of your repositories.

The full API description file is here: [github.yml](github.yml).

## Github Webhook

### Setup the Webhook

[Setup a
webhook](https://docs.github.com/en/developers/webhooks-and-events/creating-webhooks)
for each of your repositories with the following parameters:

* **Payload URL**: URL that point to your genapid instance, with `/github` as path.
* **Content type**: application/json
* **Secret**: A random string
* **SSL verification**: you can put genapid behind a web server to handle SSL. See this [conf file](../../ansible/templates/apache.conf) for an example of Apache conf.
* **Which events would you like to trigger this webhook?**: Send me everything.
* **Active**: enabled

### Match Github Webhook

We start a pipe with a filter on the path set above, using a
[`match`  predicate](../../predicates/match/):

``` yaml
- name: "Incoming github request"
  pipe:
  - name: github dedicated endpoint
    match:
      string: =In.Req.URL.Path
      value: /github
```

In order to authenticate the Webhook we have to [calculate a hash on
the
payload](https://docs.github.com/en/developers/webhooks-and-events/securing-your-webhooks). We
first get the content of the body and check that the content-type is
the one expected, using a [`body` predicate](../../predicates/body/):

``` yaml
  - name: Get body
    body:
      mime: application/json
      type: string
    register: body
```

We store the result in `R.body`.

The hash to check is passed in a header that looks like:

```
X-Hub-Signature-256: sha256=03550210cfcc0002e56e3bb72b5b695ea552ad6da8b14b31ac27590e8f791f16
```

We calculate the hash on the payload using the secret set on Github
and check it corresponds to the one passed in Github header with the
[`header` predicate](../../predicates/header/).


``` yaml
  - name: Check GitHub hash
    header:
      name: X-Hub-Signature-256
      value: '= "sha256=" + hmacSha256("GITHUB_SECRET"", R.body.payload)'
```

All parameter values that starts with an equal signs are
[expressions](../../README.md#expressions) evaluated before the
predicate is evaluated.

Now that we are sure we are processing a valid Github Webook, we parse the
JSON payload and retrieve the name of the event:

``` yaml
  - name: Parse JSON body
    body:
      type: json
    register: body
  - header:
      name: X-GitHub-Event
    register: event
```

We can then process each event we are interesting in and use its
payload. We start a new sub-pipe for each event. First, we match the
event name, here a
[ping event](https://docs.github.com/en/developers/webhooks-and-events/webhook-events-and-payloads#pi).
Then we write the zen string (inspirational phrase) in genapid log,
and use [Pushbullet API](https://docs.pushbullet.com/#create-push) to
receive a notification.

``` yaml
  - name: ping event
    pipe:

    - match:
        string: '=R.event.value'
        value: ping

    - log:
        msg: =format("received ping, %v", R.body.payload.zen)

    - http:
        url: https://api.pushbullet.com/v2/pushes
        method: post
        headers:
          Access-Token: 'PUSHBULLET_SECRET'
        body:
          json:
            type: note
            title: Github ping
            body: =R.body.payload.zen
```

The [`http` predicate](../../predicates/http/) is used to call an
external API. The *PUSHBULLET_SECRET* can be retrieved on your
[Pushbullet account
settings](https://docs.pushbullet.com/#api-quick-start) page.

To process other events, we add other sub-pipes. For example, to
receive a notification when a [pull request
event](https://docs.github.com/en/developers/webhooks-and-events/webhook-events-and-payloads#pull_request)
with the `opened` action is received:

``` yaml
  - name: pull_request event
    pipe:
    - match:
        string: '=R.event.value'
        value: pull_request
    - match:
        string: =R.body.payload.action
        value: opened
    - http:
        url: https://api.pushbullet.com/v2/pushes
        method: post
        headers:
          Access-Token: 'PUSHBULLET_SECRET'
        body:
          json:
            type: note
            title: github pull request
            body: >
              =format("received pull_request on %v from %v",
              R.body.payload.repository, R.body.payload.sender.login)
```

## Default parameters

In these examples, some parameters are repeated in several
predicates. We can factorize those parameters in a [`default`
predicate](../../README.md#default) before our sub-pipes:

``` yaml
  - default:
      http: # pushbullet push API
        url: https://api.pushbullet.com/v2/pushes
        method: post
        headers:
          Access-Token: 'PUSHBULLET_SECRET'
      match: # match event name per default
        string: '=R.event.value'
```

Now, the above event processing predicates can be written:

``` yaml
  - name: ping event
    pipe:
    - match:
        value: ping
    - log:
        msg: =format("received ping, %v", R.body.payload.zen)
    - http:
        body:
          json:
            type: note
            title: github ping
            body: =R.body.payload.zen

  - name: pull_request event
    pipe:
    - match:
        value: pull_request
    - match:
        string: =R.body.payload.action
        value: opened
    - http:
        body:
          json:
            type: note
            title: github pull request
            body: >
              =format("received pull_request on %v from %v",
              R.body.payload.repository, R.body.payload.sender.login)
```

## Voice notification on the Google Home

You can use the [`chromecast` predicate](../../predicates/chromecast/)
to use your Google Home to receive notifications.

Add the configuration for the Google Home and TTS API service with a
`default` predicate before the event sub-pipes:

``` yaml
  - default:
      chromecast: # Used to give voice notifications
        # Google Home IP addr
        addr: 192.168.3.9
        # Credentials to use Google API service for TTS
        google_service_account: "path_to/google_service_account.json"
```

Then in each event sup-pipe, you can use:

``` yaml
      - chromecast:
          tts: >
            =format("Github pull request received on %v",
            R.body.payload.repository)
```

## Local Github repository mirror

To keep an up-to-date local mirror of some Github repositories, you
can checkout the repositories in a local directory (here
`/var/lib/genapid/github/`) and use a
[`command`](../../predicates/command/) predicate in the *push event*
pipe to get new commits each time someone push something on
Github:

``` yaml
    - name: Pull commits
      command:
        cmd: git
        args:
          - pull
        chdir: >
          = "/var/lib/genapid/github/"+ R.body.payload.repository.name
```



See [github.yml](github.yml) for the complete API file description.
