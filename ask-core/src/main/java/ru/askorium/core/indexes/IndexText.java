package ru.askorium.core.indexes;

import lombok.AllArgsConstructor;
import lombok.Data;

@Data
@AllArgsConstructor
public class IndexText {

    private String key;

    private String text;

    private Integer rank;

}
