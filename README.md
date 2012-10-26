# Grout

_Under development; is technically usable, but nothing at all is set in stone or considered complete._

Grout is a static site generator written in Go, and yes, inspired by Jekyll. Frankly, this came about because of my lack of Ruby knowledge, and my desire to extend Jekyll.

### Extensible
Great extensibility is the main goal. Yet, this doesn't happen through plugins. Import Grout to make your own static site generator. It started this way because Go does not yet provide dynamic loading of code, but you'll see this design can be beneficial.

### Independent
Whatever environment your site ends up being generated on, it will only need your static site generator executable. Not the Go distribution, and not even Grout itself.

