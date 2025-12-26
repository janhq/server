import { codeToHtml } from 'shiki';
import { CodeExampleClient } from './code-example-client';

const examples = [
  {
    title: 'Python',
    language: 'python',
    code: `from Jan import Jan

client = Jan()  

response = client.responses.create(
  model="jan-v1",
  input="Tell me a three sentence bedtime story about a unicorn"
)

print(response.output_text)`,
  },
  {
    title: 'JavaScript',
    language: 'javascript',
    code: `import Jan from "Jan";

const client = new Jan();

const response = await client.responses.create({
  model: 'jan-v1',
  input: 'Tell me a three sentence bedtime story about a unicorn'
});

console.log(response.output_text);`,
  },
  {
    title: 'cURL',
    language: 'bash',
    code: `curl https://platform.jan.ai/v1/responses \\
-H "Authorization: Bearer $Jan_API_KEY" \\
-H "Content-Type: application/json" \\
-d '{
  "model": "jan-v1",
  "input": "Tell me a three sentence bedtime story about a unicorn"
}'`,
  },
];

export async function CodeExample() {
  // Pre-render syntax highlighting on the server
  const highlightedExamples = await Promise.all(
    examples.map(async (example) => {
      const html = await codeToHtml(example.code, {
        lang: example.language,
        themes: {
          light: 'github-light',
          dark: 'github-dark',
        },
        defaultColor: false,
      });
      return {
        ...example,
        highlightedHtml: html,
      };
    }),
  );

  return <CodeExampleClient examples={highlightedExamples} />;
}
