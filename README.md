# jsondescriber
A Go package for generating inventories, comparisons, and English-language descriptions of arbitrary JSON data.

Most useful for comparing and sanity-checking version changes to static JSON objects, such as config files.

Conceived for the use case of generating warning and confirmation dialogs for an AWS SecretsManager wrapper, e.g.: are you sure you want to send a bare string instead of a structured object; are you about to overwrite dozens of entries with just one or two.
