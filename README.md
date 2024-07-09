# Find Your Next Job or Ideal Candidate with AI-Powered Hacker News Search

This demo showcases how to use vector search and Large Language Models (LLMs) to intelligently search Hacker News "Who is Hiring" posts based on job queries or candidate profiles.

## For Job Seekers:
Simply provide your resume or a link to your LinkedIn profile, and we'll leverage AI to find suitable job opportunities from Hacker News.

## For Employers:
Submit a resume or LinkedIn profile of your ideal candidate, and we'll identify similar candidates from the Hacker News talent pool.

## How It Works:
1. We load all posts and top-level comments from "whoishiring" posts into a local SQLite database.
2. Embeddings are generated for each post using one of the supported embedding models.
3. When given a job query or candidate profile:
   a. The LLM suggests relevant search terms.
   b. We search the database for comments matching these terms using the pre-calculated embeddings.
   c. Comments are ranked based on embedding similarity to the search terms.
   d. Similar comments from the same user are removed to ensure diversity.
   e. The LLM provides recommendations based on the top K comments.

## Supported Embedding Models:
* Ollama: gemini:2, nomic-embed-text
* VoyageAI: voyage-2
* OpenAI: text-embedding-3-small

**Note:** Testing has shown that Voyage or OpenAI embeddings work best.

## Supported Completion Models:
* Anthropic: Claude
* OpenAI: GPT models

## Configuration:
Set the following environment variables:
- `OPENAI_API_KEY`
- `ANTHROPIC_API_KEY`
- `VOYAGE_API_KEY`
- `PROXYCURL_API_KEY`

**Note:** proxycurl is used to scrape LinkedIn profiles.

No worries if you don't have all the keys, the demo will still work.

## Usage:
Select the embedding model:
```
-embedding=openai|nomic-embed-text|gemma:2b|text-embedding-3-small|voyage-2
```

Select the completion model:
```
-completion=claude|openai
```

Use cached results (for testing):
```
-fake=true|false
```

## Default settings:
- Embedding model: voyage-2
- Completion model: claude