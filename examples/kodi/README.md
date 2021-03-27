# Controlling Kodi with Google Assistant

genapid can be used to control [Kodi](https://kodi.tv/) by voice using a Google Home, like with [GoogleHomeKodi](https://github.com/OmerTu/GoogleHomeKodi#how-to-setup-and-update), and use the Google Assistant to give you the results of your actions.

A full example of an API configuration file is available in [kodi.yml](kodi.yml).

## IFTTT

We use IFTTT to receive the phrase from Google Assistant, as described [here](../../doc/google-assistant.md) and [here](../../doc/ifttt.md).

After the first predicates that check we are processing a valid request from IFTTT, we use a sub-pipe for each command we want to process.

## Mute sound

The phrase received by Google Assistant is passed in the URL query parameter we've setup in the Webhook in IFTTT, which is `q`.

In genapid, the URL query parameters are available in `In.URL.Params`.

We use a [`match`](../../predicates/match/) predicate to match the phrase against a regexp.

Note that the `name` options are optional and only used for documenting and logs readability.

``` yaml
  - name: Mute sound
    pipe:
    - name: Match a phrase asking to mute sound
      match:
        string: = In.URL.Params.q[0]
        regexp:  ^(mute|unmute)( the sound)?$
```


The `match` predicate is true if the regexp matched our query. In this case, the next predicate is evaluated. We use a [`jsonrpc`](../../predicates/jsonrpc/) predicate to call the [Kodi JSON-RPC API](https://kodi.wiki/view/JSON-RPC_API):

``` yaml
    - name: Mute sound on Kodi
      jsonrpc:
        procedure: Application.SetMute
        params:
          mute: toggle
        url: http://192.168.0.32:8080/jsonrpc
        basic_auth:
          username: kodi
          password: mykodipasswd
```

## Play a movie

For the next action, we start a new `pipe`. We first match the phrase like before, but with a named group in the regexp to get the movie title. We register the results in `movie`.

``` yaml
  - name: Play movie
    pipe:
    - name: Match a phrase asking to play a movie
      match:
        string: = In.URL.Params.q[0]
        regexp: "play (the movie )?(?P<title>.+)"
      register: movie
```

We call Kodi to get the list of all movies and store the results in `movies`:

``` yaml
    - name: Get the list of all available movies
      jsonrpc:
        procedure: VideoLibrary.GetMovies
      register: movies
```

We use a jsonpath expression to get the list of titles in the `titles` variable from the result of the jsonrpc call:

``` yaml
    - name: Get the list of movies
      set:
        - titles:  '= R.movies.response | $.movies[*].label'
```

We do a fuzzy search to find the best matching title, using the result of the regexp `match` in the `titles` variable set before:

``` yaml
    - name: get best matching title
      set:
        - title:   '= fuzzy(R.movie.named.title, V.titles)'
```

And we use a jsonpath to get the id of the matched movie:

``` yaml
    - name: get id corresponding to matched title
      set:
        - movieid: '= jsonpath(format(`$.movies[?(@.label=="%s")].movieid`, V.title), R.movies.response)'
```

Now we can ask Kodi to play the movie using its ID. Note that a jsonpath filter always returns a list.

``` yaml
    - name: Play the movie
      jsonrpc:
        procedure: Player.Open
        params:
          item:
            movieid: "=V.movieid[0]"
        url: http://192.168.0.32:8080/jsonrpc
        basic_auth:
          username: kodi
          password: mykodipasswd
```

## Default parameters

In these examples, some parameters are repeated in several predicates. We can factorize those parameters by using a `default` predicate at the beginning of our main pipe:

``` yaml
  - name: Set default values
    default:
      match: # Phrase from Google Assistant sent by IFTTT
        string: = In.URL.Params.q[0]
      jsonrpc: # Kodi API params
        url: http://192.168.0.32:8080/jsonrpc
        basic_auth:
          username: kodi
          password: mykodipasswd
```

Now, the above mute action can be written:
``` yaml
  - name: Mute sound
    pipe:
    - name: Match a phrase asking to mute sound
      match:
        regexp:  ^(mute|unmute)( the sound)?$

    - name: Mute sound on Kodi
      jsonrpc:
        procedure: Application.SetMute
        params:
          mute: toggle
```

## Voice feedback on the Google Home

You can use the [`chromecast`](../../predicates/chromecast/) predicate to use your Google Home to give you feedbacks on your actions.

For example, if the movie you asked to play is not found, you can use this predicate just before the jsonrpc call to `Player.Open` (default values are omitted, see [complete example](kodi.yml)):

``` yaml
    - name: Movie not found
      chromecast:
        tts: '=format("No movie named %v", R.movie.named.title)'
      when: '=len(V.movieid) == 0'
      result: =false # stop here
```

## Other actions

Check [kodi.yml](kodi.yml) API config file for more examples.
