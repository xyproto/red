package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/PullRequestInc/go-gpt3"
	"github.com/xyproto/env/v2"
	"github.com/xyproto/mode"
	"github.com/xyproto/vt100"
)

const (
	codePrompt     = "Write it in %s and include comments where it makes sense. The code should be concise, correct and expertly created. Comments above functions should start with the function name."
	continuePrompt = "Write the next 10 lines of this %s program:\n"
	textPrompt     = "Write it in %s. It should be expertly written, concise and correct."
)

// ProgrammingLanguage returns true if the current mode appears to be a programming language (and not a markup language etc)
func (e *Editor) ProgrammingLanguage() bool {
	switch e.mode {
	case mode.Blank, mode.AIDL, mode.Amber, mode.Bazel, mode.Config, mode.Doc, mode.Email, mode.Git, mode.HIDL, mode.HTML, mode.JSON, mode.Log, mode.M4, mode.ManPage, mode.Markdown, mode.Nroff, mode.PolicyLanguage, mode.ReStructured, mode.Shader, mode.SQL, mode.Text, mode.XML:
		return false
	}
	return true
}

// AIFixups adds a space after single-line comments
func (e *Editor) AIFixups(generatedLine string) string {
	singleLineComment := e.SingleLineCommentMarker()
	trimmedLine := strings.TrimSpace(generatedLine)
	if len(trimmedLine) > 2 && e.ProgrammingLanguage() && strings.HasPrefix(trimmedLine, singleLineComment) && !strings.HasPrefix(trimmedLine, singleLineComment+" ") && !strings.HasPrefix(generatedLine, "#!") {
		return strings.Replace(generatedLine, singleLineComment, singleLineComment+" ", 1)
	}
	return generatedLine
}

// GenerateTokens uses the ChatGTP API to generate text. n is the maximum number of tokens.
// The global atomic Bool "ContinueGeneratingTokens" controls when the text generation should stop.
func (e *Editor) GenerateTokens(apiKey, prompt string, n int, temperature float32, model string, newToken func(string)) error {
	client := gpt3.NewClient(apiKey)
	chatContext, cancelFunction := context.WithCancel(context.Background())
	defer cancelFunction()
	err := client.CompletionStreamWithEngine(
		chatContext,
		model,
		gpt3.CompletionRequest{
			Prompt:      []string{prompt},
			MaxTokens:   gpt3.IntPtr(n),
			Temperature: gpt3.Float32Ptr(temperature),
		}, func(resp *gpt3.CompletionResponse) {
			newToken(resp.Choices[0].Text)
			if !e.generatingTokens {
				cancelFunction()
			}
		})
	return err
}

// TODO: Find an exact way to find the number of tokens in the prompt, from a ChatGPT point of view
func countTokens(s string) int {
	// Multiplying with 1.1 and adding 100, until the OpenAI API for counting tokens is used
	return int(float64(len(strings.Fields(s)))*1.1 + 100)
}

// GenerateCodeOrText will try to generate and insert text at the corrent position in the editor, given a ChatGPT prompt
func (e *Editor) GenerateCodeOrText(c *vt100.Canvas, status *StatusBar, bookmark *Position, chatAPIKey, chatPrompt string) {
	if chatAPIKey == "" {
		status.SetErrorMessage("ChatGPT API key is empty")
		status.Show(c, e)
		return
	}

	// Strip away any leading exclamation marks and trim away spaces at the ends
	prompt := strings.TrimSpace(strings.TrimSuffix(chatPrompt, "!"))

	const (
		generateText = iota
		generateCode
		continueCode
	)

	generationType := generateText // generateCode // continueCode
	if e.ProgrammingLanguage() {
		generationType = generateCode
		if prompt == "" {
			generationType = continueCode
		}
	}

	// Determine the temperature
	var defaultTemperature float32
	switch generationType {
	case generateText:
		defaultTemperature = 0.8
	}
	temperature := env.Float32("CHATGPT_TEMPERATURE", defaultTemperature)

	// Select a model
	//gptModel, gptModelTokens := gpt3.TextDavinci003Engine, 4000
	gptModel, gptModelTokens := "gpt-4", 4000
	// gptModel, gptModelTokens := "gpt-3.5-turbo", 4000 // only for chat
	// gptModel, gptModelTokens := "text-curie-001", 2048 // simpler and faster
	// gptModel, gptModelTokens := "text-ada-001", 2048 // even simpler and even faster

	switch generationType {
	case continueCode:
		gptModel, gptModelTokens = "code-davinci-002", 8000
		// gptModel, gptModelTokens = "code-cushman-001", 2048 // slightly simpler and slightly faster
	}

	// Prefix the prompt
	switch generationType {
	case generateCode:
		prompt += ". " + fmt.Sprintf(codePrompt, e.mode.String())
	case continueCode:
		prompt += ". " + fmt.Sprintf(continuePrompt, e.mode.String()) + "\n"
		// gather about 2000 tokens/fields from the current file and use that as the prompt
		startTokens := strings.Fields(e.String())
		gatherNTokens := gptModelTokens - countTokens(prompt)
		if len(startTokens) > gatherNTokens {
			startTokens = startTokens[len(startTokens)-gatherNTokens:]
		}
		prompt += strings.Join(startTokens, " ")
	case generateText:
		prompt += ". " + fmt.Sprintf(textPrompt, e.mode.String())
	}

	// Set a suitable status bar text
	status.ClearAll(c)
	switch generationType {
	case generateText:
		status.SetMessage("Generating text...")
	case generateCode:
		status.SetMessage("Generating code...")
	case continueCode:
		status.SetMessage("Continuing code...")
	}
	status.Show(c, e)

	// Find the maxTokens value that will be sent to the OpenAI API
	amountOfPromptTokens := countTokens(prompt)
	maxTokens := gptModelTokens - amountOfPromptTokens // The user can press Esc when there are enough tokens
	if maxTokens < 1 {
		status.SetErrorMessage("ChatGPT API request is too long")
		status.Show(c, e)
		return
	}

	// Start generating the code/text while inserting words into the editor as it happens
	currentLeadingWhitespace := e.LeadingWhitespace()
	e.generatingTokens = true // global
	first := true
	var generatedLine string
	if err := e.GenerateTokens(chatAPIKey, prompt, maxTokens, temperature, gptModel, func(word string) {
		generatedLine += word
		if strings.HasSuffix(generatedLine, "\n") {
			e.SetCurrentLine(currentLeadingWhitespace + e.AIFixups(generatedLine))
			if !first {
				if !e.EmptyTrimmedLine() {
					e.InsertLineBelow()
					e.pos.sy++
				}
			} else {
				e.DeleteCurrentLineMoveBookmark(bookmark)
				first = false
			}
			generatedLine = ""
		} else {
			e.SetCurrentLine(currentLeadingWhitespace + e.AIFixups(generatedLine))
		}
		// "refresh"
		e.MakeConsistent()
		e.DrawLines(c, true, false)
		e.redrawCursor = true
	}); err != nil {
		e.redrawCursor = true
		errorMessage := err.Error()
		if !strings.Contains(errorMessage, "context") {
			e.End(c)
			status.SetError(err)
			status.Show(c, e)
			return
		}
	}
	e.End(c)

	if e.generatingTokens { // global
		if first { // Nothing was generated
			status.SetMessageAfterRedraw("Nothing was generated")
		} else {
			status.SetMessageAfterRedraw("Done")
		}
	} else {
		status.SetMessageAfterRedraw("Stopped")
	}

	e.RedrawAtEndOfKeyLoop(c, status)
}
