export namespace models {
	
	export class AIAnalysis {
	    tags: string[];
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new AIAnalysis(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tags = source["tags"];
	        this.description = source["description"];
	    }
	}
	export class AIConfig {
	    id: number;
	    apiKey: string;
	    model: string;
	    baseURL: string;
	    enabled: boolean;
	    enableAgent: boolean;
	    enableWebSearch: boolean;
	    maxTokens: number;
	    timeout: number;
	
	    static createFrom(source: any = {}) {
	        return new AIConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.apiKey = source["apiKey"];
	        this.model = source["model"];
	        this.baseURL = source["baseURL"];
	        this.enabled = source["enabled"];
	        this.enableAgent = source["enableAgent"];
	        this.enableWebSearch = source["enableWebSearch"];
	        this.maxTokens = source["maxTokens"];
	        this.timeout = source["timeout"];
	    }
	}
	export class Metadata {
	    size?: number;
	    width?: number;
	    height?: number;
	    duration?: number;
	    pageCount?: number;
	    author?: string;
	    title?: string;
	
	    static createFrom(source: any = {}) {
	        return new Metadata(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.size = source["size"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.duration = source["duration"];
	        this.pageCount = source["pageCount"];
	        this.author = source["author"];
	        this.title = source["title"];
	    }
	}
	export class File {
	    id: number;
	    fileName: string;
	    originalName: string;
	    filePath: string;
	    fileType: string;
	    fileSubType: string;
	    fileSize: number;
	    tags: string[];
	    description: string;
	    metadata: Metadata;
	    thumbnail: string;
	    checksum: string;
	    isDeleted: boolean;
	    deletedAt: string;
	    createdAt: string;
	    modifiedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new File(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.fileName = source["fileName"];
	        this.originalName = source["originalName"];
	        this.filePath = source["filePath"];
	        this.fileType = source["fileType"];
	        this.fileSubType = source["fileSubType"];
	        this.fileSize = source["fileSize"];
	        this.tags = source["tags"];
	        this.description = source["description"];
	        this.metadata = this.convertValues(source["metadata"], Metadata);
	        this.thumbnail = source["thumbnail"];
	        this.checksum = source["checksum"];
	        this.isDeleted = source["isDeleted"];
	        this.deletedAt = source["deletedAt"];
	        this.createdAt = source["createdAt"];
	        this.modifiedAt = source["modifiedAt"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FileType {
	    id: number;
	    name: string;
	    displayName: string;
	    extensions: string;
	    subTypes: string[];
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new FileType(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.displayName = source["displayName"];
	        this.extensions = source["extensions"];
	        this.subTypes = source["subTypes"];
	        this.createdAt = source["createdAt"];
	    }
	}
	
	export class StorageDir {
	    id: number;
	    path: string;
	    fileType: string;
	    currentSize: number;
	    maxSize: number;
	    isActive: boolean;
	    isDefault: boolean;
	    parentId?: number;
	    createdAt: string;
	    totalSize?: number;
	    subDirs?: StorageDir[];
	
	    static createFrom(source: any = {}) {
	        return new StorageDir(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.path = source["path"];
	        this.fileType = source["fileType"];
	        this.currentSize = source["currentSize"];
	        this.maxSize = source["maxSize"];
	        this.isActive = source["isActive"];
	        this.isDefault = source["isDefault"];
	        this.parentId = source["parentId"];
	        this.createdAt = source["createdAt"];
	        this.totalSize = source["totalSize"];
	        this.subDirs = this.convertValues(source["subDirs"], StorageDir);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ThemeConfig {
	    id: number;
	    themeMode: string;
	    primaryColor: string;
	    backgroundColor: string;
	    backgroundImage: string;
	
	    static createFrom(source: any = {}) {
	        return new ThemeConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.themeMode = source["themeMode"];
	        this.primaryColor = source["primaryColor"];
	        this.backgroundColor = source["backgroundColor"];
	        this.backgroundImage = source["backgroundImage"];
	    }
	}

}

