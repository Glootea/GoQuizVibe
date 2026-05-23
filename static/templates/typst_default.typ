// Typst Demo Document


// Document settings
#set page(width: 210mm, height: 297mm, margin: 2cm)
#set text(font: "Linux Libertine", size: 12pt)

// Heading style
#show heading: it => {
  text(18pt, weight: "bold", it.body)
}

// Main title
= Document Title

Lorem ipsum dolor sit amet. Body text with *emphasis* and **strong** text.

== Section One

Subsection content with bullet list:
- First item
- Second item
- Third item

== Section Two

=== Code Example

```python
def hello():
    print("Hello, World!")
```

=== Math Example

Formula: $$E = mc^2$$

Equation:

$ sum_(k=0)^n k
    &= 1 + ... + n \
    &= (n(n+1)) / 2 $


== Section Three

Nested list:
1. Numbered item
2. Another item
  - Sub-item
  - Another sub-item

> Block quote text here