lingo-toml
==========

Very basic Golang library for i18n.

The original project was by [@kortemy](https://github.com/kortemy/lingo), and uses JSON. A fork by [jdkeke142](https://github.com/jdkeke142/lingo-toml) changed the input to TOML. This fork abstracts the file system away, allowing the use of (for example) embedded assets, or assets that don't live on the local disk. It was put together around [broccoli](https://github.com/aletheia-icu/broccoli), with which it's been tested.

Features:
---------
1. Storing messages in TOML files.
2. Support for nested declarations.
3. Detecting language based on Request headers.
4. Very simple to use.

Usage:
------
  1. Import Lingo into your project

      ```go
        import "github.com/xxxserxxx/lingo"
      ```
  1. Create a dir to store translations, and write them in TOML files named [locale].toml. For example:

      ```    
        en_US.toml
        sr_RS.toml
        de.toml
        ...
      ```
      You can write nested TOML too.
      ```toml
        title = "CutleryPlus"
        subtitle = "Knives that put cut in cutlery."

          [menu]
          home = "Home"

            [menu.products]
            self = "Products"
            forks = "Forks"
            knives = "Knives"
            spoons = "Spoons"
      ```
  2. Initialize a Lingo like this:

      ```go
        l := lingo.New("default_locale", "path/to/translations/dir", nil)
      ```

      This is where you would pass in a `lingo.FileSystem` if you want lingo to read from something other than the disk. Passing in `nil` is the same as calling:

      ```go
        l := lingo.New("default_locale", "path/to/translations/dir", lingo.OSFS())
      ```

  3. Get bundle for specific locale via either `string`:

      ```go
        t1 := l.TranslationsForLocale("en_US")
        t2 := l.TranslationsForLocale("de_DE")
      ```
      This way Lingo will return the bundle for specific locale, or default if given is not found.
      Alternatively (or primarily), you can get it with `*http.Request`:

      ```go
        t := l.TranslationsForRequest(req)
      ```
      This way Lingo finds best suited locale via `Accept-Language` header, or if there is no match, returns default.
      `Accept-Language` header is set by the browser, so basically it will serve the language the user has set to his browser.
  4. Once you get T instance just fire away!

      ```go
        r1 := t1.Value("main.subtitle")
        // "Knives that put cut in cutlery."
        r1 := t2.Value("main.subtitle")
        // "Messer, die legte in Besteck geschnitten."
        r3 := t1.Value("menu.products.self")
        // "Products"
        r5 := t1.Value("error.404", req.URL.Path)
        // "Page index.html not found!"
      ```

Behaviors
---------

-  Lingo will return the default locale if it can't find a requested locale.
-  Lingo will return the default locale's text if it can't find a translation for the key in the requested locale.
-  Lingo will return the requested *key* if it can't find either the requested or the default text.
-  Text values can have positional tags, e.g. `{0}`, `{1}`.  These will be replaced by any varargs supplied to `Value`.  E.g.:

    ```
    thing="Hey there, {1}, how is {0}?"
    ```

    and

    ```go
    res := t1.Value("thing", "your dog", "person")
    ```

    will produce

    ```
    Hey there, person, how is your dog
    ```
-  Locales **must** match the file names.  E.g., `en_US` must be in a file called `en_US.toml`, not `en_us.toml`.
-  Files in the translation directory that do not end in `.toml` are ignored.


Contributions
-------------
This was forked to support another project; I'll accept PRs and will fix bugs, but am unlikely to add features.
