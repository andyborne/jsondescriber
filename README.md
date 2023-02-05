# jsondescriber
A Go package for generating inventories and English-language descriptions of arbitrary JSON data.

May be useful for comparing version changes.

Conceived for the use case of generating warning and confirmation dialogs for an AWS SecretsManager wrapper, i.e. are you sure you want to send a bare string instead of an object, or are you about to overwrite dozens of entries with just one or two.
