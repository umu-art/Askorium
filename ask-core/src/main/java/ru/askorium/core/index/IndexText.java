package ru.askorium.core.index;

import lombok.AllArgsConstructor;
import lombok.Data;

@Data
@AllArgsConstructor
public class IndexText {

    private String key;

    private String text;

    private Float rank;

}
