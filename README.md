Goat Web Framework
==================

Goat is a small web framework for Go that makes the following assumptions about your application:

* It is using a Mongodb database
* It has a User model that has a username and password

Goat wraps the stock Go net/http library to provide a faster API for building routes for your application.
Routes can have universally configured middleware that will execute on each request, or an interceptor
that will only fire on specific requests.

# Basic usage

Add Goat to your project by importing the package

        import "github.com/vokalinteractive/goat"

Then initialize it within your application

        g := goat.NewGoat()

You'll probably want sessions and a database in your app, so configure the middleware as well

        g.RegisterMiddleware(g.NewSessionMiddleware("mysessionstore"))
	    g.RegisterMiddleware(g.NewDatabaseMiddleware("localhost", "database_name"))

Now that you're middleware is configured you can specify your routes, as well as any interceptors that you
want to use

        g.RegisterRoute("/login", goat.GET|goat.POST, "login", Login)
        g.RegisterRoute("/", goat.GET, "index", goat.NewAuthSessionInterceptor(GetUserDetails, Login))
        g.RegisterStaticFileHandler("/static/")

And you're ready to start serving your app

        g.ListenAndServe("8080")

# Routes

Routes adhere to the Handler type:

        func(w http.ResponseWriter, r *http.Request, c *goat.Context)

The context attached to each request is how you can access your database, session, or user from within a
view handler

        type Context struct {
            Database *mgo.Database
            Session  *sessions.Session
            User     *User
        }

# Middleware

Middleware is a function that returns a function with the following signature:

        func(r *http.Request, c *goat.Context)

All middleware will run in the order in which it is registered, before a handler is called on your route.
Out of the box, Goat provides the following middleware:

        // Enables session management on your requests
        NewSessionMiddleware(storename string) Middleware

        // Configures a database connection and clones that connect onto each request
        NewDatabaseMiddleware(host, name string) Middleware

# Interceptor

Sometimes you may want requests to perform an action that other requests shouldn't do. Since middleware isn't a
viable solution in this case, Goat provides interceptors. An interceptor returns a function with a Handler
type:

        func(w http.ResponseWriter, r *http.Request, c *goat.Context)

Goat provides the following interceptors for you:

        // Authenticates a request via basic auth
        NewBasicAuthInterceptor(normal Handler) Interceptor

        // Verifies that the session associated with this request has a goat.User associated with it,
        // otherwise it redirects to unauthorized
        NewAuthSessionInterceptor(normal, unauthorized Handler) Interceptor
