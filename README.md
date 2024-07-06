This is a demo of how to use vector search and LLM to search hacker news "who is hiring" posts
based on a job query prompt.

All posts and top level comments from posts by "whoishiring" are loaded into a local sqlite database.

We then generate embeddings from the text using one of the supported embeddings -- either from ollama
using gemini:2, nomic-embed-text, or openai text-embedding-3-small. Testing has shown that the openai
embeddings are better.

The search is done as follows:
1. Given a job query prompt, we ask the LLM for suggested search terms.
2. We then search the database for comments that match the search terms given the precalculated embeddings.
3. We then rank the comments based on the similarity of the embeddings to the search terms.
4. We then remove any similar comments from the same user.
5. We then ask the LLM for recommendations based on the top k comments.

The demo supports both openai and claude.

The keys must be set in environment variables: OPENAI_API_KEY, ANTHROPIC_API_KEY

To select the models to use in main.go:

```
embeddingModel  = OpenAI3Small
completionModel = Claude
```

