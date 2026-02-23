package ru.askorium.core.ask_scrapper_api.service;

import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import ru.askorium.api.model.ScrappedPage;
import ru.askorium.core.ask_scrapper_api.AskScrapperService;

import java.util.List;

@Service
@RequiredArgsConstructor
public class AskScrapperServiceImpl implements AskScrapperService {

    @Override
    public List<ScrappedPage> scrapSource(String sourceUrl) {
        return List.of(); // TODO: complete
    }

}
