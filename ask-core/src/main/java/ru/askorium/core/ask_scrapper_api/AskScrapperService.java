package ru.askorium.core.ask_scrapper_api;

import ru.askorium.api.model.ScrappedPage;

import java.util.List;

public interface AskScrapperService {

    List<ScrappedPage> scrapSource(String sourceUrl);

}
