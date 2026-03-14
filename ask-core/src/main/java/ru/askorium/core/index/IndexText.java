package ru.askorium.core.index;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class IndexText {

    private String key;

    private String text;

    private Float rank;

}
