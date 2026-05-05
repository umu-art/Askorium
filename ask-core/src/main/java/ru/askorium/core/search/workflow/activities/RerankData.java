package ru.askorium.core.search.workflow.activities;

import lombok.AllArgsConstructor;
import lombok.Data;

import java.util.List;
import java.util.Map;

@Data
@AllArgsConstructor
public class RerankData {
    private String query;
    private List<Map<String, Object>> blocks;
}
