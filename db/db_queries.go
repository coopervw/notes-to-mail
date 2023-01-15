package db

const GetAllBooksDbQueryConstant = `
	select 
		ZBKLIBRARYASSET.ZASSETID,
		ZBKLIBRARYASSET.ZTITLE,
		ZBKLIBRARYASSET.ZAUTHOR,    
		count(a.ZAEANNOTATION.Z_PK)
	from ZBKLIBRARYASSET left join a.ZAEANNOTATION
		on a.ZAEANNOTATION.ZANNOTATIONASSETID = ZBKLIBRARYASSET.ZASSETID
	WHERE a.ZAEANNOTATION.ZANNOTATIONSELECTEDTEXT NOT NULL
	GROUP BY ZBKLIBRARYASSET.ZASSETID;
`

const GetBookDataById = `
	select
		ZBKLIBRARYASSET.ZTITLE,
		ZBKLIBRARYASSET.ZAUTHOR
	from ZBKLIBRARYASSET
	where ZBKLIBRARYASSET.ZASSETID=$1
`

const GetNotesHighlightsById = `
	select 
		a.ZAEANNOTATION.ZANNOTATIONSELECTEDTEXT,
		a.ZAEANNOTATION.ZANNOTATIONNOTE
	from
		a.ZAEANNOTATION
	where 
		a.ZAEANNOTATION.ZANNOTATIONASSETID = $1
		AND a.ZAEANNOTATION.ZANNOTATIONSELECTEDTEXT NOT NULL
	order by ZPLLOCATIONRANGESTART ASC, ZANNOTATIONCREATIONDATE ASC
`
