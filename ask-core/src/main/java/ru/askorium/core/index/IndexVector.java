package ru.askorium.core.index;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class IndexVector {

    private String key;

    private List<Float> values;

    private Float rank;

}
