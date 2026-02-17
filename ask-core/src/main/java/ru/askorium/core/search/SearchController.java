package ru.askorium.core.search;

import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Controller;
import ru.askorium.api.server.SearchApi;
import ru.askorium.api.server.model.SearchCreateRequest;
import ru.askorium.api.server.model.SearchCreateResponse;
import ru.askorium.api.server.model.SearchGetResponse;

import java.util.UUID;

@Controller
@RequiredArgsConstructor
public class SearchController implements SearchApi {

    @Override
    public ResponseEntity<SearchCreateResponse> createSearchQuery(SearchCreateRequest searchCreateRequest) {
        return null;
    }

    @Override
    public ResponseEntity<SearchGetResponse> getSearchQueryResult(UUID queryId) {
        return null;
    }

}
