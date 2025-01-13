package templates

var IntentInstruction = `
  You are given an instruction along description of task labelled as [Task Description]. For the given instruction, list out 3-5 keywords in comma separated format as [Intent] which define the characteristics or properties required by the about the most capable and suitable agent to solve the task using the instruction.


  [Task Description]: {{task_description}}
  [Instruction]: {{instruction}}
  
  
  [Intent]:
`
