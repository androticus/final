# final-ai AI Chat provider

final-ai provides a web application that provides an AI chat interface. Under the hood, it creates an AI tool that the AI can use to help satisfy user requests.

To use final-ai:

final-ai -port port -apikey key -provider provider

port: default to 8088
apikey: required, env default: FINAL_AI_API_KEY
provider: default is 'grok' others: 'chatgpt'

When the user browses to the eg localhost:8080 they receive a web interface that is a simple version of contemporary AI chat interfaces.

At the bottom is a text box that expands vertically as the user types or pastes in text. At the right is a prominent button with test or an icon to indicate send.

When the user sends a new prompt, the chat history of the session is bundled up and sent to the AI. Along with this, a description of Final as a tool will also be sent. The AI can use the Final session to help satisfy user requests. It will also allow users to enter the definitions of new Final words, in Final syntax, to be added to the Session.

Above the text box at the bottom is a scrollable history of user prompts and AI replies. The user prompts should be set off, such as in a differently styled text box, not spanning the whole width, to the right. This is the customary chat interface. The AI reply to a given prompt is then show full width below the user prompt display box.

## Implementation

The code should be written using the commonly provided ChatGPT api, which is also accepted by Grok. The first version only works with Grok.

The user prompts and AT replies should be included in a context value, which is appended for each chat turn. The format of this should follow current chat AI conventions.

An AI tool will be described for the AI. It will provide a succinct summary of Final, that it is similar to Forth, but extends it. The APIs to provide the tool creation services will be in package ai in internal/ai, and exposed to the embedding Library.

Tool requests will be delegated to a suitable interface on the Session object. As a high level description, it will accept a string of text that is Final interpetor/compiler source, send it to be interpreted by the Session instance of Final, then the result of the interpretion session will be returned. This will enable the AI to run existing Final words in interpetive mode, eg `1 2 +` or define new : words, which it can subsequently invoke in the session string afer definition.

The web page application should be kept fairly simple, and not require complex additional go packages to implement. However, a small curated set of new dependencies that will minimize the amount of new code that has to be generated will be welcome.

Implementing this application will require a fair bit of new go code. Each slice of the implementation plan will describe the proposed changes and additions, and the proposed local of new or revised code in these possible places: package internal/core, package internal/ai, or application final-ai.

The abiity to use Final in the manner demonstrated by this application is a PRIMARY USE CASE, so we need to pay attention to code factoring, and end-user library friendliness. Ideally, it should require too many lines of code in a user application to add Final as a tool for its AI.
