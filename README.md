# go-map

Go's 1.19 hashmap implementation using pure Go with generics.

## About

This is Go's 1.19 map implementation written with pure Go. Using this repo you can learn how maps work in Go internally and undestand the main concept of a hashmap in general. This repo uses algorithms from an actual map implementation under the Go's hood, except some things due to better undestanding.

This code was written additionaly to my [article]() on Habr, where I've desribed maps internals - what is a hashmap, main terms, concepts, difference with Python and Java.

[Here](https://prog.world/hashmap-according-to-golang-along-with-implementation-on-generics/) you can find some kind of auto-generated English translation.

You can also use this repo as a start point to improve/change a base implementation that we have in Go 1.19.

# Contributions

Any contributions are welcome. Don't hesitate creating PRs to improve similarity to a base implementation, to fix bugs, typos, redability etc.

# Links

Articles\videos which will help you to delve into a hashmap:

Go:
[GopherCon 2016: Keith Randall - Inside the Map Implementation](https://www.youtube.com/watch?v=Tl7mi9QmLns)
[How the Go runtime implements maps efficiently (without generics)](https://dave.cheney.net/2018/05/29/how-the-go-runtime-implements-maps-efficiently-without-generics)

Python:
[Raymond Hettinger Modern Python Dictionaries A confluence of a dozen great ideas PyCon 2017 ](https://www.youtube.com/watch?v=npw4s1QTmPg)[Raymond Hettinger.
More compact dictionaries with faster iteration](https://mail.python.org/pipermail/python-dev/2012-December/123028.html)

Java:
[The Java HashMap Under the Hood
](https://www.baeldung.com/java-hashmap-advanced)[Liner probing lecture. cs166 stanford](https://web.stanford.edu/class/archive/cs/cs166/cs166.1166/lectures/12/Small12.pdf)
[An Analysis of Hash Map Implementations in Popular Languages](https://rcoh.me/posts/hash-map-analysis/)
