## A **Prompt Orchestrator** with conditional modules

This gives you dynamic control at runtime:

* add memory or not
* add tools or not
* add templates or patterns or not
* customize tone or behavior
* assemble final prompt automatically

---

A **Prompt Orchestrator** is a custom logic layer (code or workflow) that:

1. Takes a user’s raw input
2. Checks conditions (flags, context, user settings, memory, etc.)
3. Composes a final prompt programmatically
4. Sends that composed prompt to an LLM Inference Engine

---

## What a Prompt Orchestrator Can Do

It can automatically attach optional modules, such as:

### Memory

If user enables memory, insert memory instructions into prompt.

### Tool usage

Conditionally include instructions like:

* “use the retrieval tool when needed”
* “use the calculator tool if numbers appear”

### Templates / prompt patterns

For example:

* Chain-of-Thought structure
* Output format
* Persona / role descriptions
* “First think step-by-step, then answer”

### Safety rules

Add system-level constraints when specific topics appear.

### Output shapers

Like “respond in JSON”, “respond concisely”, “use a teacher tone”, etc.

### Conditional behaviors

* If question is about code → add code assistant template
* If question mentions “summarize” → add summary template
* If user speaks Vietnamese → switch language automatically

---

## HOW TO BUILD a Prompt Orchestrator

You can build it in any backend (Node.js, Python, etc.).
The architecture is simple:

```
User Input
    ↓
Prompt Orchestrator
    - Check context
    - Check rules
    - Retrieve memory
    - Assemble components
    ↓
Final Prompt (System + Instructions + User)
    ↓
LLM Inference API
```

---

## Step 1 — Define Prompt Modules

Example modules you can mix-and-match:

### **Base System Prompt**

```
You are a helpful assistant. Follow the rules strictly.
```

### **Memory Module**

```
Use the following personal memory for this user:
{{memory}}
```

### **Tool Instructions Module**

```
You have access to the following tools: {{tools}}
Always choose the best tool for the task.
```

### **Style / Persona Module**

```
Respond in friendly tone unless user asks otherwise.
```

### **Task Templates**

* writing template
* analysis template
* translation template
* technical breakdown template

---

## Step 2 — Build Conditional Logic

Example (pseudocode):

```python
prompt = BASE_SYSTEM_PROMPT

if use_memory:
    prompt += MEMORY_MODULE.replace("{{memory}}", retrieved_memory)

if question_is_code:
    prompt += CODE_ASSISTANT_TEMPLATE

if user_language == "vi":
    prompt += VIETNAMESE_STYLE_TEMPLATE

if use_tools:
    prompt += TOOL_INSTRUCTIONS_MODULE
```

---

## Step 3 — Output Combined Prompt

Example combined output:

```
You are a helpful assistant.

Use the following memory for this user:
- wife prefers female voice
- avoid parentheses in Mermaid diagrams

Respond in a structured style:
1. Explanation
2. Output
3. Notes

User request:
"How do I build a pricing model for my SaaS?"
```

Then you send this to Inference LLM (system + user).

---