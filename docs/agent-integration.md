# Agent Integration Guide

## Overview

LocalSpace now uses intelligent agents for automatic tag and description generation during file import. This guide explains how the agent system works and how to use it effectively.

## Architecture

The agent system consists of several components:

- **AgentService**: Manages tag and description generation agents
- **TagAgent**: Specialized agent for generating file tags
- **DescriptionAgent**: Specialized agent for generating file descriptions
- **WebSearchTool**: Provides web search capability for agents
- **PromptBuilder**: Constructs intelligent prompts for AI generation

## Features

### 1. Keyword Input

Users can now provide keywords when importing files to guide AI generation:

```go
req := ImportFileRequest{
    FilePath:    "/path/to/file.mp4",
    FileName:    "movie.mp4",
    Keywords:    "action thriller movie",
    Tags:        []string{"entertainment"},
    Description: "An action thriller movie",
}
```

### 2. Intelligent Generation

Agents analyze multiple sources of information:
- File name and type
- User-provided keywords
- User-provided tags
- User-provided description
- File metadata

### 3. Autonomous Web Search

Agents can automatically decide when to search the web for additional information:
- Unfamiliar software names
- Game titles
- Complex filenames
- Limited user input

### 4. Error Resilience

AI generation failures never block file imports:
- Graceful degradation to basic generation
- User input is always preserved
- Errors are logged but don't interrupt the process

## Configuration

### AI Configuration

Extend your AI configuration to enable agent features:

```json
{
  "ai": {
    "enabled": true,
    "apiKey": "your-api-key",
    "model": "gpt-4",
    "baseURL": "https://api.openai.com/v1",
    "enableAgent": true,
    "enableWebSearch": true,
    "maxTokens": 500,
    "timeout": 30
  }
}
```

### Configuration Options

- `enabled`: Enable/disable AI features globally
- `enableAgent`: Enable intelligent agent generation
- `enableWebSearch`: Allow agents to search the web
- `maxTokens`: Maximum tokens for AI responses
- `timeout`: Timeout in seconds for AI operations

## Usage

### Basic Import

```go
req := ImportFileRequest{
    FilePath: "/path/to/file.pdf",
    FileName: "document.pdf",
    // No keywords, tags, or description - agent will generate
}

err := fileService.ImportFile(req)
```

### Import with User Input

```go
req := ImportFileRequest{
    FilePath:    "/path/to/file.pdf",
    FileName:    "document.pdf",
    Keywords:    "business report finance",
    Tags:        []string{"work", "important"},
    Description: "Annual financial report",
}

err := fileService.ImportFile(req)
```

### Programmatic Generation

```go
ctx := context.Background()

// Generate tags
tags, err := agentService.GenerateTags(
    ctx,
    "movie.mp4",
    "video",
    "action thriller",
    []string{"entertainment"},
    "An action movie",
)

// Generate description
description, err := agentService.GenerateDescription(
    ctx,
    "document.pdf",
    "document",
    "business",
    []string{"work"},
    "Annual report",
)
```

## Best Practices

### 1. Provide Context When Possible

Give the agents more information to work with:

```go
// Good
req := ImportFileRequest{
    FileName:     "software-installer.exe",
    Keywords:     "productivity tool office suite",
    Tags:         []string{"software", "installer"},
}

// Less effective
req := ImportFileRequest{
    FileName: "installer.exe",
}
```

### 2. Use Descriptive Filenames

Descriptive filenames help agents understand file content:

```
Good: Annual_Financial_Report_2024.pdf
Less effective: document.pdf
```

### 3. Leverage User Tags

Provide initial tags to guide generation:

```go
req := ImportFileRequest{
    FileName: "game-setup.exe",
    Tags:     []string{"game", "installer"},
    Keywords: "RPG adventure game",
}
```

### 4. Monitor Performance

Keep an eye on AI usage and costs:

- Check logs for generation errors
- Monitor API call frequency
- Adjust timeout settings if needed

## Troubleshooting

### AI Generation Not Working

1. Check AI configuration is enabled
2. Verify API key is valid
3. Check network connectivity
4. Review logs for error messages

### Poor Quality Results

1. Provide more keywords or user input
2. Use more descriptive filenames
3. Enable web search for more context
4. Adjust AI model or temperature settings

### Slow Performance

1. Reduce timeout values
2. Disable web search if not needed
3. Use faster AI model
4. Consider caching results

## Performance Expectations

- **Tag generation**: < 3 seconds (without web search)
- **Tag generation**: < 8 seconds (with web search)
- **Description generation**: < 3 seconds (without web search)
- **Description generation**: < 8 seconds (with web search)
- **Concurrent requests**: Support for 10+ simultaneous operations

## Security Considerations

- User keywords and tags are never logged
- Web search queries are anonymized
- API keys are stored securely
- Input validation prevents injection attacks

## Future Enhancements

Planned improvements to the agent system:

- [ ] Batch generation for multiple files
- [ ] Custom prompt templates
- [ ] User feedback learning
- [ ] Vector database for semantic search
- [ ] Local model support

## Support

For issues or questions about agent integration:

1. Check this documentation first
2. Review logs for error messages
3. Verify configuration settings
4. Test with simple examples first