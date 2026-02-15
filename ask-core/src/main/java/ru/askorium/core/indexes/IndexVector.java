package ru.askorium.core.indexes;

import lombok.AllArgsConstructor;
import lombok.Data;

import java.util.List;

@Data
@AllArgsConstructor
public class IndexVector {

    private String key;

    private List<Float> values;

    private Integer rank;

}
