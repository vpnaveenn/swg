# Go Character Normalization Tool (`charNorm`)

This tool generates string variations based on Unicode character normalizations. Its primary purpose is to aid in testing for Cross-Site Scripting (XSS) and Server-Side Request Forgery (SSRF) vulnerabilities by creating payloads that might bypass naive security filters.

The normalization data is fetched from an HTML table located at: `https://raw.githubusercontent.com/zy9ard3/zy9ard3.github.io/refs/heads/main/normalizations.html`.

## Modules

The project consists of two main Go modules:

1.  **`example.com/charNorm`**: This is the core library module.
    *   It handles the parsing of the normalization table.
    *   It provides the logic for generating string variations.
    *   Key files:
        *   `parser.go`: Contains the `ParseNormalizationTable(url string) (map[rune][]rune, error)` function, which fetches and parses the HTML normalization table.
        *   `generator.go`: Contains the `GenerateVariations(input string, normMap map[rune][]rune) []string` function, which produces the character variations of an input string.

2.  **`example.com/charNormCLI`**: This is the command-line interface (CLI) tool that utilizes the `example.com/charNorm` library.
    *   It allows users to provide an input string (via command-line argument or stdin) and get the generated variations.
    *   Key file:
        *   `cmd/charNormCLI/main.go`: The main application entry point for the CLI tool.

## Building the CLI

To build the `charNormCLI` tool:

1.  Navigate to the CLI's module directory:
    ```bash
    cd cmd/charNormCLI
    ```
2.  Build the executable:
    ```bash
    go build -o charNormCLI
    ```
    This command compiles the `main.go` program and creates an executable file named `charNormCLI` (or `charNormCLI.exe` on Windows) within the `cmd/charNormCLI` directory.

## Running the CLI

Once built, you can run the `charNormCLI` from within the `cmd/charNormCLI` directory.

**Using command-line arguments:**

Provide the string you want to generate variations for as a command-line argument. If your string contains spaces or special characters, enclose it in quotes.

```bash
./charNormCLI "<script>alert(1)</script>"
```

**Using standard input (stdin):**

You can pipe input directly to the CLI tool. This is useful for scripting or when working with output from other commands.

```bash
echo "http://localhost/admin" | ./charNormCLI
```

In both cases, the tool will output the list of generated payload variations to standard output, with each variation on a new line.

## Example Usage for Bug Hunting

The generated string variations are intended to be used in security testing, particularly for identifying XSS and SSRF vulnerabilities.

*   The generated variations can be copied and pasted directly into web application inputs, such as search bars, form fields, or URL parameters.
*   They can be integrated with security testing tools like Burp Suite Intruder, OWASP ZAP's fuzzer, or custom scripts. This allows for automated testing of how a web application handles these normalized forms compared to standard ASCII input. The goal is to discover if the application's input validation or sanitization routines can be bypassed by using these Unicode equivalents.

**Illustrative Example:**

If you provide the input "cat":

Input:
```bash
./charNormCLI "cat"
```

Example Output (the actual output will be significantly longer, this is just a small sample):
```
cat
ĉat
ḉat
cȧt
cḁt
caŧ
caƭ
... (many more variations)
```
This output can then be used to test if an application that expects "cat" behaves differently or becomes vulnerable when presented with "ĉat", "cȧt", etc.
