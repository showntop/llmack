package templates

// ExampleGenerate 样本生成提示词
const ExampleGenerate = `
  You are an expert example selector who can help in selection of right in-context examples to help the most suitable agent solve this problem.
  You are also given the prompt instruction which is used to solve this task
  [Prompt]: {{prompt}}
  You are given the task description of the task:
  [Task Description]: {{task_description}}
  I'm trying to write a few shots prompt using {{num_examples}} in-context examples to effectively solve any questions of the above task.
  Think of analysing, understanding and creating examples of task on the criteria of diversity of types of examples, complexity of the nature/characteristics of the examples and relevance/compatibility to the whole example set in total.
  Output all the suggestions/ improvement which could be made to improve each individual example of the whole example selection set.
`

// ExampleOptimization 样本优化提示词
const ExampleOptimization = `
  You are an expert example selector who can help in selection of right in-context examples to help the agent solve this problem.
  You are also given the prompt instruction which is used to solve this task
  [Prompt]: {prompt}
  You are given the description of the task:
  [Task Description]: {task_description}
  I'm trying to write a few shots prompt using {num_examples} in-context examples to effectively solve any questions of the above task.
  My current {num_examples} in-context examples set are: {examples}
  You are also given a set of suggestions/improvements which could be made to improve each individual example of the whole example selection set:
  [SUGGESTION/IMPROVEMENT]: {critique}
  Based on the above information, use all of it smartly and diligently to carefully create new set of {num_examples}, which follow these suggestion and improvements.
  Make sure to output each example wrapped with <START> and <END>.
  
  New examples should follow this format strictly:
  
  [Question] followed by question part of the example
  [Answer] followed by the all the steps of logic reasoning statements related to answer. The final answer as "<ANS_START>[answer]<ANS_END>"
  
  For Example: <START>
  {gt_example}
  <END>
  
  [New Examples]:
  `
