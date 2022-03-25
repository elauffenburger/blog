---
title: "Language Constraints"
date: 2022-03-18T01:16:51-07:00
draft: true
---

# Constraints in Language Design

I've been thinking a while about what makes a good language good (or, maybe even better, a _bad_ one bad), and here're a few things that I've come up with:

	* Expressiveness
	* Safety
	* Disambiguity
	* Self-Consistency
	* Ease-of-use
	* Some other stuff I'm forgetting

I've heard this described as a language's set of "values" before -- the things that drive idioms, language features, and generally give a language the "feel" that it has. It's...a bit hard to describe, but once you've work with a language for a while, you kinda start to understand what would or wouldn't be idiomatic, and I think that comes directly from these values!

This is all pretty abstract, so let's take a look at some examples!

## Examples In Different Languages

Here's the same program translated to a few different languages so we can get a feel for each one's values. I'm going with a problem that isn't really supposed to be a comprehensive look at what a language can or can't do, but rather what kind of work each one is designed for! I think (and maybe this is controversial!) most modern languages are pretty much identical in terms of what you can do with them. That doesn't mean they have the same feature sets (obviously! Otherwise, this would be a short post...), but they're all roughly capable of getting the same result. Some are going to be more elegant than others because of these tradeoffs we're talking about, but I'd be _really_ shocked if there was a language out there that couldn't do what we're about to try.

Okay, so for reference, here's the problem we're going to solve in a couple different languages:

```
Given a string like "aaabccdddd", compress the representation to "a3bc2d4".
```

Okay, so I'm officially not the best at writing interview prompts, but this is a pretty good interview question I've seen a couple times in phone screens! I like it for a couple reasons:
	* It's pretty straightforward to understand (I won't say "easy" -- I could write a whole post on that...)
	* It doesn't require anything magickal 🧙‍♂️
	* You can solve it in pretty much any language ever

So, let's dig into it!

### JavaScript

First off, let's do this in JavaScript. *dodges tomatoes* Okay look, so I know it's not everyone's favorite, but I'd argue that similar to how C was the lingua franca of software development for a _long_ time, that crown (for...better or worse) has fallen to JS. And that's not a bad thing, honest!

I was recently helping a bio-engineering student out with his CS homework and these maniacs had him writing a little CLI program in C++ that involved `const` references and `vector`s and `char` -> ASCII and -- well, you get the idea!

Oh, and this was week 10.

...week 10.

If you want to do this to CS or CSE students at week 10, fine; we did this to ourselves! But what is _any_ of that teaching engineering students? I'd argue it's definitely not teaching them critical thinking or problem solving -- and I'd argue heavily that _no one_ is ready at week 10 to start talking about pass-by-value vs pass-by-reference. Maybe there's someone reading this right now going "psh, that's like _day_ 10; no, wait: _hour_ 10, you scrubs!", but I'm just going to ignore them for now and let 'em go back to shitposting about their galaxy-sized brain on Reddit or whatever.

Wait, where was I? Oh yeah, programming and stuff.

So, anyways, here's some JavaScript:

```js
function encode(str) {
	if (!str) {
		return '';
	}

	let result = '';
	let lastChar = str[0];
	let charCount = 1;

	for (let i = 1; i < str.length; i++) {
		const c = str[i];
		if (c == lastChar) {
			charCount++;

			continue;
		}

		result += charCount > 1 
			? `${lastChar}${charCount}`
			: lastChar

		lastChar = c;
		charCount = 1;
	}

	return result;
}
```

Not super exciting, and not super pretty either. You could totally  code-golf ⛳ this, but that's not really the point! I think everything here is pretty idiomatic and looks like JS you might see out in the wild.

Okay, so let's take some notes:


### TypeScript

## Expressiveness

	* What lets you impart your artistic vision on the world
	* Part of that Secret Sauce™ that enables elegant solutions

## Safety

	* I think of this as being "anything the compiler does to prevent you from shooting yourself in the foot later"
	* Call me a greybeard or whatever, but (for 👉 this guy 👈 at least) that usually means strong, static typing _at the very least_ [^2]

## Disambiguity

	* The language's ability to give you the the tools to be very very explicit about what you're doing to prevent confusion.

## Self-Consistency

	* The way a language feels "properly put together"
		* Thoughtfulness
		* A shared vision
		* No obvious design inconsistencies
		* There's probably a better word for this whole thing...

[^1]: Actually, on that note, if you know of any language communities (or communities in general!) that started out with one set of values but ended up pivoting to another one, shoot me a message! I _seriously_ would love to hear about it!
[^2]: What's stronger than that, you say?? [SPARK](https://en.wikipedia.org/wiki/SPARK_(programming_language)), [Eiffel](https://en.wikipedia.org/wiki/Eiffel_(programming_language)), and their ilk would like a word with you.
