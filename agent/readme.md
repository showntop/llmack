
# CoT
论文：https://arxiv.org/pdf/2201.11903.pdf
传统的 Prompt 从输入直接到输出的映射 < input——>output > 的方式
CoT 完成了从输入到思维链再到输出的映射，即 < input——>reasoning chain——>output >
所以本质上是希望模型把思考的过程暴露出来，达到可解释、可控的目的，COT应该是大模型本身已经具备的能力，我们只是使用方式唤醒它。

仅仅在 Prompt 中添加了一句 "Let's Step by Step" 就让大模型在推理上用到了思维链。

CoT的自我一致性，多次迭代取多数结果作为最终结果

# ToT
CoT的缺陷：
- 对于局部，没有探索一个思考过程下的不同延续-树的分支。
- 对于全局，没有利用任何类型的规划，前瞻以及回溯去帮助评估不同抉择-而启发式的探索正式人类解决问题的特性。

论文：https://arxiv.org/pdf/2305.10601.pdf

![ToT示意](../../docs/assets/ToT.png)
核心是根据思考过程发散成树形结构，类似于思维导图，不断扩散推理过程。

# ReAct
论文：https://arxiv.org/pdf/2210.03629.pdf。
Cot和Tot都是专注于大模型本身的推理，ReAct则是整个系统的统筹。从推理过程，结合外部工具共同实现最终的目标。
Agent 最常用的实现思路之一，它强调在执行任务时结合推理（Reasoning）和行动（Acting）两个方面，使得Agent能够在复杂和动态的环境中更有效地工作。