# Google Assistant / Google Home

To receive a command from Google Assistant, we can use [IFTT](https://ifttt.com/).

With the setup below, you have to say something like this:
> OK Google Home, *do something*

### This
For **This** choose: **Google Assistant**

1. Choose *Say a phrase with a text ingredient*
2. In *What do you want to say?* enter something like:
> home $

3. In *What do you want the Assistant to say in response?* enter something like:
> ok

Then setup a [*That* Webhook](ifttt.md) to receive the query on genapid.
