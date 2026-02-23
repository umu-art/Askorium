package ru.askorium.core.search.workflow;

import io.temporal.workflow.WorkflowInterface;
import io.temporal.workflow.WorkflowMethod;

import java.util.UUID;

@WorkflowInterface
public interface SearchWorkflow {

    @WorkflowMethod
    void search(UUID queryId);

}
