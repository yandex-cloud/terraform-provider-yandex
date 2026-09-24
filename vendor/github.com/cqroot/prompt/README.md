<div align="center">
  <h1>✔ Prompt</h1>
  <p>
    User-friendly, highly customizable interactive prompts for Go.
    <br />
    Based on <a href="https://github.com/charmbracelet/bubbletea" alt="Bubble Tea">Bubble Tea</a>.
    Inspired by <a href="https://github.com/terkelg/prompts" alt="Prompts">Prompts</a>
      and <a href="https://github.com/charmbracelet/gum" alt="Gum">Gum</a>.
  </p>

  <p>
    <a href="https://github.com/cqroot/prompt/actions">
      <img src="https://github.com/cqroot/prompt/workflows/test/badge.svg" alt="Action Status" />
    </a>
    <a href="https://codecov.io/gh/cqroot/prompt">
      <img src="https://codecov.io/gh/cqroot/prompt/branch/main/graph/badge.svg" alt="Codecov" />
    </a>
    <a href="https://goreportcard.com/report/github.com/cqroot/prompt">
      <img src="https://goreportcard.com/badge/github.com/cqroot/prompt" alt="Go Report Card" />
    </a>
    <a href="https://pkg.go.dev/github.com/cqroot/prompt">
      <img src="https://pkg.go.dev/badge/github.com/cqroot/prompt.svg" alt="Go Reference" />
    </a>
    <a href="https://github.com/cqroot/prompt/tags">
      <img src="https://img.shields.io/github/v/tag/cqroot/prompt" alt="Git tag" />
    </a>
    <a href="https://github.com/cqroot/prompt/blob/main/go.mod">
      <img src="https://img.shields.io/github/go-mod/go-version/cqroot/prompt" alt="Go Version" />
    </a>
    <a href="https://github.com/cqroot/prompt/blob/main/LICENSE">
      <img src="https://img.shields.io/github/license/cqroot/prompt" />
    </a>
    <a href="https://github.com/cqroot/prompt/issues">
      <img src="https://img.shields.io/github/issues/cqroot/prompt" />
    </a>
  </p>
</div>

## Table of Contents

- [Features](#features)
- [Screenshots](#screenshots)
  - [Choose](#choose)
  - [AdvancedChoose](#advancedchoose)
  - [MultiChoose](#multichoose)
  - [Input](#input)
  - [Write](#write)
  - [Prompt Theme](#prompt-theme)
- [Customization](#customization)
- [License](#license)

## Features

1. `choose` lets the user choose one of several strings using the terminal ui.
2. `multichoose` lets the user choose multiple strings from multiple strings using the terminal ui.
3. `input` lets the user enter a string using the terminal ui.
   You can specify that only **numbers** or **integers** are allowed.
4. Show help message for keymaps.
5. Based on [Bubble Tea]("https://github.com/charmbracelet/bubbletea").
   `prompt.Prompt` and all child models implement `tea.Model`.

## Screenshots

### Choose

[example](https://github.com/cqroot/prompt/blob/main/_examples/choose/main.go)

![choose](https://user-images.githubusercontent.com/46901748/219288366-d4ce04df-ca98-4a03-8a80-e7c26577e86a.gif)

**Modify the theme of choose:**

[example](https://github.com/cqroot/prompt/blob/main/_examples/choose-themes/main.go)

![choose-themes](https://user-images.githubusercontent.com/46901748/219293300-cb1cd6ac-d43f-414f-b526-f490423b7108.gif)

### AdvancedChoose

[example](https://github.com/cqroot/prompt/blob/main/_examples/advancedchoose/main.go)

![advancedchoose](https://user-images.githubusercontent.com/46901748/232372136-ac5fc71f-25bb-4379-979a-f0e16a7058d8.gif)

### MultiChoose

[example](https://github.com/cqroot/prompt/blob/main/_examples/multichoose/main.go)

![multichoose](https://user-images.githubusercontent.com/46901748/219288777-1c913ac8-4144-4b96-b5be-3085483d8bae.gif)

**Modify the theme of choose:**

[example](https://github.com/cqroot/prompt/blob/main/_examples/multichoose-themes/main.go)

![multichoose-themes](https://user-images.githubusercontent.com/46901748/219293895-137d82f6-7344-4ea0-aa34-85110aaa9c0d.gif)

### Input

[example](https://github.com/cqroot/prompt/blob/main/_examples/input/main.go)

![input](https://user-images.githubusercontent.com/46901748/219288988-12923602-a112-4876-906d-3575f3c50741.gif)

**Password input**

[example](https://github.com/cqroot/prompt/blob/main/_examples/input-echo-password/main.go)

![input-echo-password](https://user-images.githubusercontent.com/46901748/218799172-ce501335-9821-4bf2-949a-0c08057d810f.gif)

**Password input like linux (do not display any characters)**

[example](https://github.com/cqroot/prompt/blob/main/_examples/input-echo-none/main.go)

![input-echo-none](https://user-images.githubusercontent.com/46901748/218799167-59b52b0d-228e-4cb3-8bf2-7cf844874100.gif)

**Only integers can be entered**

[example](https://github.com/cqroot/prompt/blob/main/_examples/input-integer-only/main.go)

**Only numbers can be entered**

[example](https://github.com/cqroot/prompt/blob/main/_examples/input-number-only/main.go)

**Input with validation**

[example](https://github.com/cqroot/prompt/blob/main/_examples/input-with-validation/main.go)

![input-with-validation](https://user-images.githubusercontent.com/46901748/218799174-9355fcb1-bcef-4fe6-8421-e9472e913010.gif)

### Write

[example](https://github.com/cqroot/prompt/blob/main/_examples/write/main.go)

![write](https://user-images.githubusercontent.com/46901748/219289253-7fef6708-c852-4d88-b2d0-376249f46c9b.gif)

### Prompt Theme

All model themes can be customized. The prompt's theme can also be customized.

[example](https://github.com/cqroot/prompt/blob/main/_examples/prompt-themes/main.go)

![prompt-themes](https://user-images.githubusercontent.com/46901748/219320761-223f9be7-bb2f-4851-9b80-5a8ebee8074d.gif)

## Customization

Some options can be passed when using these models, such as whether to display help information, etc.

All available options and examples can be seen in the following files:

- Choose [options](https://github.com/cqroot/prompt/blob/main/choose/options.go).
- MultiChoose [options](https://github.com/cqroot/prompt/blob/main/multichoose/options.go), [example](https://github.com/cqroot/prompt/blob/main/_examples/multichoose-options/main.go).
- Input [options](https://github.com/cqroot/prompt/blob/main/input/options.go), [example](https://github.com/cqroot/prompt/blob/main/_examples/input-options/main.go).
- Write [options](https://github.com/cqroot/prompt/blob/main/write/options.go), [example](https://github.com/cqroot/prompt/blob/main/_examples/write-options/main.go).

## License

[MIT License](https://github.com/cqroot/prompt/blob/main/LICENSE).
