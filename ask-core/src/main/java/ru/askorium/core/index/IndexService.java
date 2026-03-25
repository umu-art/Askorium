package ru.askorium.core.index;

import java.util.List;

public interface IndexService {

    void saveVectors(List<IndexVector> vectors);

    void saveTexts(List<IndexText> texts);

    List<IndexText> searchBM25(String query, int size);

    List<IndexVector> searchKnn(IndexVector vector, int size);

    void deleteStaleKeys(List<String> validKeys);

}
