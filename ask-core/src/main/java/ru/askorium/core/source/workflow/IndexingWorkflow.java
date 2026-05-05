package ru.askorium.core.source.workflow;

import io.temporal.workflow.WorkflowInterface;
import io.temporal.workflow.WorkflowMethod;

import java.util.UUID;

@WorkflowInterface
public interface IndexingWorkflow {

    @WorkflowMethod
    void index(UUID pageId);

}
