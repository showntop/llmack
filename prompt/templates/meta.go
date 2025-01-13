package templates

// MetaInstruction is the instruction for generating prompts with meta prompts.
var MetaInstruction = `
  You are given a task description and a prompt instruction and different styles known as meta prompts:
  [Task Description]: {{task_description}}
  [Meta Prompt]: {{meta_prompts}}
  Now you need to generate {{num_variations}} variations of following Instruction adaptively mixing meta prompt while keeping similar semantic meaning.
  Make sure to wrap each generated prompt with <START> and <END>
  [Prompt Instruction]: {{prompt_instruction}}
  [Generated Prompts]:
`
